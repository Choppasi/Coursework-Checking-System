package middleware

import (
	"net/http"
	"sync"
	"time"
)

type client struct {
	count    int
	lastSeen time.Time
}

var (
	clients = make(map[string]*client)
	mu      sync.RWMutex
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		mu.Lock()
		c, exists := clients[ip]
		if !exists || time.Since(c.lastSeen) > time.Minute {
			clients[ip] = &client{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		c.count++
		c.lastSeen = time.Now()
		count := c.count
		mu.Unlock()

		if count > 30 {
			http.Error(w, `{"error":"Too many requests"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
