package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientData
}

type clientData struct {
	count    int
	lastSeen time.Time
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{clients: make(map[string]*clientData)}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Limit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ClientIP(r)

			rl.mu.Lock()
			c, exists := rl.clients[ip]
			if !exists || time.Since(c.lastSeen) > window {
				rl.clients[ip] = &clientData{count: 1, lastSeen: time.Now()}
				rl.mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			c.count++
			c.lastSeen = time.Now()

			if c.count > maxRequests {
				rl.mu.Unlock()
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}
			rl.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > 10*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}
