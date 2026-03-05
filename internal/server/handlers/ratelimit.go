package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RateLimiter implements a simple token bucket rate limiter per IP address.
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int           // Maximum number of requests
	window   time.Duration // Time window for rate limiting
	stop     chan struct{}
}

// NewRateLimiter creates a new rate limiter with the specified limit and window.
// limit: Maximum number of requests allowed
// window: Time window (e.g., 1 minute)
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		stop:     make(chan struct{}),
	}

	// Start cleanup goroutine to remove old entries
	go rl.cleanupExpired()

	return rl
}

// Allow checks if a request from the given IP address should be allowed.
// Returns true if the request should be allowed, false if rate limit exceeded.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this IP
	requests, exists := rl.requests[ip]
	if !exists {
		requests = []time.Time{}
	}

	// Remove expired requests (older than the window)
	validRequests := []time.Time{}
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we've exceeded the limit
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests

	return true
}

// cleanupExpired periodically removes expired entries to prevent memory leaks.
func (rl *RateLimiter) cleanupExpired() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mutex.Lock()
			now := time.Now()
			cutoff := now.Add(-rl.window)

			for ip, requests := range rl.requests {
				validRequests := []time.Time{}
				for _, reqTime := range requests {
					if reqTime.After(cutoff) {
						validRequests = append(validRequests, reqTime)
					}
				}

				if len(validRequests) == 0 {
					delete(rl.requests, ip)
				} else {
					rl.requests[ip] = validRequests
				}
			}
			rl.mutex.Unlock()
		case <-rl.stop:
			return
		}
	}
}

// Stop stops the cleanup goroutine. Should be called when the rate limiter is no longer needed.
// Safe to call multiple times.
func (rl *RateLimiter) Stop() {
	// Only close if not already closed
	select {
	case <-rl.stop:
		// Already closed
	default:
		close(rl.stop)
	}
}

// RateLimitMiddleware creates a Gin middleware that rate limits requests based on IP address.
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()

		if !limiter.Allow(clientIP) {
			log.Warn().
				Str("client_ip", clientIP).
				Str("path", ctx.Request.URL.Path).
				Str("method", ctx.Request.Method).
				Msg("Rate limit exceeded")

			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":             "too_many_requests",
				"error_description": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		ctx.Next()
	}
}
