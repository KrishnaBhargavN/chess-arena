package auth

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// clientIP returns the caller's IP address. Behind Cloudflare the real client
// address arrives in CF-Connecting-IP, and since we're only reachable through
// the tunnel that header is trustworthy. Locally it falls back to the socket.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// bucket is a token bucket that refills at a steady rate up to a cap.
type bucket struct {
	tokens   float64
	lastFill time.Time
}

// RateLimiter throttles requests per client IP, one token bucket each.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	refill   float64 // tokens added per second
	capacity float64 // max tokens, i.e. the burst size
}

// NewRateLimiter allows up to `burst` requests at once, then one more every
// 1/refillPerSec seconds. Idle entries are swept periodically.
func NewRateLimiter(refillPerSec float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		refill:   refillPerSec,
		capacity: float64(burst),
	}
	go rl.sweep()
	return rl
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok {
		rl.buckets[ip] = &bucket{tokens: rl.capacity - 1, lastFill: now}
		return true
	}

	b.tokens += now.Sub(b.lastFill).Seconds() * rl.refill
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastFill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) sweep() {
	for range time.Tick(10 * time.Minute) {
		rl.mu.Lock()
		for ip, b := range rl.buckets {
			if time.Since(b.lastFill) > 10*time.Minute {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit wraps a handler, rejecting callers that exceed their rate with 429.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r)) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
