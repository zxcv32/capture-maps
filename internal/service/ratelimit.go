package service

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// A single /print request can fan out to dozens of outbound, billable Google
// Maps API calls, so each client IP gets its own token bucket: a small burst
// for normal use, then throttled to one request every rateLimitInterval.
const (
	rateLimitInterval = 5 * time.Second
	rateLimitBurst    = 3
)

var (
	limiterMu sync.Mutex
	limiters  = make(map[string]*rate.Limiter)
)

// ponytail: limiters never evicts entries, so it grows one key per unique
// client IP for the life of the process. Fine at hobby-project traffic; add
// a TTL sweep (e.g. delete entries idle >10m) if this sees real abuse volume.
func limiterFor(ip string) *rate.Limiter {
	limiterMu.Lock()
	defer limiterMu.Unlock()
	l, ok := limiters[ip]
	if !ok {
		l = rate.NewLimiter(rate.Every(rateLimitInterval), rateLimitBurst)
		limiters[ip] = l
	}
	return l
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimit throttles each client IP independently so a single caller can't
// run up the owner's Google Maps bill by hammering the handler it wraps.
// Preflight OPTIONS requests are let through unthrottled since they carry no
// print work and the handler already short-circuits them.
func RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		if !limiterFor(clientIP(r)).Allow() {
			http.Error(w, "rate limit exceeded, try again shortly", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
