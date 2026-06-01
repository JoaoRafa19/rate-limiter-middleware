package api

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipLimiter struct {
	mu       sync.Mutex
	r        rate.Limit
	b        int
	visitors map[string]*visitor
}

func newIpLimiter(r rate.Limit, b int) *ipLimiter {
	lim := &ipLimiter{
		visitors: make(map[string]*visitor),
		r:        r,
		b:        b,
	}

	go lim.Cleanup()

	return lim
}

func (l *ipLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.visitors[ip]

	if !ok {
		lim := rate.NewLimiter(l.r, l.b)
		visitor := &visitor{
			limiter:  lim,
			lastSeen: time.Now(),
		}
		l.visitors[ip] = visitor
		return lim
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (l *ipLimiter) Cleanup() {
	for {
		time.Sleep(time.Minute)
		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastSeen) > time.Minute*3 {
				delete(l.visitors, ip)
			}
		}

		l.mu.Unlock()

	}
}

func clientIp(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func IPRateLimiter(l *ipLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.getLimiter(clientIp(r)).Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				w.Header().Set("Retry-After", "5")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
