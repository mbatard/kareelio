package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultPort       = "8080"
	defaultBackendURL = "http://backend:8080"
	defaultDistDir    = "/app/dist"
)

func main() {
	port := env("PORT", defaultPort)
	backendURL := env("BACKEND_URL", defaultBackendURL)
	distDir := env("DIST_DIR", defaultDistDir)

	if _, err := os.Stat(filepath.Join(distDir, "index.html")); err != nil {
		log.Fatalf("missing frontend dist: %v", err)
	}

	target, err := url.Parse(backendURL)
	if err != nil {
		log.Fatalf("invalid BACKEND_URL: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/readyz", readyz)
	mux.Handle("/api/", newProxy(target))
	mux.HandleFunc("/", spaHandler(distDir))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           securityHeaders(mux),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("frontend server starting on :%s backend_url=%s dist_dir=%s", port, backendURL, distDir)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("frontend server failed: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func readyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func newProxy(target *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, `{"error":"bad gateway"}`, http.StatusBadGateway)
	}
	return proxy
}

func spaHandler(distDir string) http.HandlerFunc {
	indexPath := filepath.Join(distDir, "index.html")
	assetServer := http.FileServer(http.Dir(distDir))
	return func(w http.ResponseWriter, r *http.Request) {
		requestPath := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		if requestPath == "/" {
			serveFile(w, r, indexPath, false)
			return
		}

		if containsHiddenSegment(requestPath) {
			http.NotFound(w, r)
			return
		}

		if filepath.Ext(requestPath) != "" {
			if isAssetFile(requestPath) {
				w.Header().Set("Cache-Control", "public, immutable, max-age=31536000")
			} else {
				w.Header().Set("Cache-Control", "no-store")
			}
			assetServer.ServeHTTP(w, r)
			return
		}

		serveFile(w, r, indexPath, false)
	}
}

func containsHiddenSegment(requestPath string) bool {
	for _, segment := range strings.Split(strings.TrimPrefix(requestPath, "/"), "/") {
		if segment == "" {
			continue
		}
		if strings.HasPrefix(segment, ".") {
			return true
		}
	}
	return false
}

func serveFile(w http.ResponseWriter, r *http.Request, filePath string, cacheable bool) {
	if cacheable {
		w.Header().Set("Cache-Control", "public, immutable, max-age=31536000")
	} else {
		w.Header().Set("Cache-Control", "no-store")
	}
	http.ServeFile(w, r, filePath)
}

func isAssetFile(filePath string) bool {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".woff", ".woff2", ".ttf", ".eot":
		return true
	default:
		return false
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		next.ServeHTTP(w, r)
	})
}
