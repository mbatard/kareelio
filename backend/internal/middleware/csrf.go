package middleware

import (
	"net/http"
	"net/url"
	"strings"
)

func CSRFProtection(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := strings.ToUpper(r.Method)
			if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
				next.ServeHTTP(w, r)
				return
			}

			if origin := r.Header.Get("Origin"); origin != "" {
				if !isAllowedOrigin(origin, origins) {
					http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if referer := r.Header.Get("Referer"); referer != "" {
				parsed, err := url.Parse(referer)
				if err != nil {
					http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
					return
				}
				refererOrigin := parsed.Scheme + "://" + parsed.Host
				if !isAllowedOrigin(refererOrigin, origins) {
					http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		})
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, a := range allowed {
		a = strings.TrimSpace(a)
		if a == "*" || origin == a {
			return true
		}
	}
	return false
}
