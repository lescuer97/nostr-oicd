package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientInfo struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients sync.Map // map[string]*clientInfo
	once    sync.Once
)

// RateLimitMiddleware returns a Chi middleware that rate-limits requests by client IP.
// rps is the allowed requests per second (use rate.Every(time.Minute/requests) for per-minute),
// burst is the allowed burst size.
func RateLimitMiddleware(rps rate.Limit, burst int) func(next http.Handler) http.Handler {
	// start cleanup once
	once.Do(func() { go cleanupStaleClients() })

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			ci := getClient(ip, rps, burst)
			ci.lastSeen = time.Now()
			if !ci.limiter.Allow() {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getClient(ip string, rps rate.Limit, burst int) *clientInfo {
	if v, ok := clients.Load(ip); ok {
		return v.(*clientInfo)
	}
	lim := rate.NewLimiter(rps, burst)
	ci := &clientInfo{limiter: lim, lastSeen: time.Now()}
	clients.Store(ip, ci)
	return ci
}

func cleanupStaleClients() {
	for {
		time.Sleep(5 * time.Minute)
		cutoff := time.Now().Add(-10 * time.Minute)
		clients.Range(func(k, v interface{}) bool {
			ci := v.(*clientInfo)
			if ci.lastSeen.Before(cutoff) {
				clients.Delete(k)
			}
			return true
		})
	}
}

// clientIP returns a best-effort client IP address.
// It checks X-Forwarded-For, X-Real-IP, and falls back to r.RemoteAddr.
func clientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// may contain multiple addresses, use first
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	// fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
