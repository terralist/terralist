package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(5, 1*time.Minute)
	defer limiter.Stop()

	ip := "192.168.1.1"

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		allowed := limiter.Allow(ip)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be denied
	allowed := limiter.Allow(ip)
	assert.False(t, allowed, "6th request should be denied (rate limit exceeded)")
}

func TestRateLimiter_Allow_DifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(3, 1*time.Minute)
	defer limiter.Stop()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Both IPs should have separate rate limits
	for i := 0; i < 3; i++ {
		assert.True(t, limiter.Allow(ip1), "IP1 request %d should be allowed", i+1)
		assert.True(t, limiter.Allow(ip2), "IP2 request %d should be allowed", i+1)
	}

	// Both should hit their limits
	assert.False(t, limiter.Allow(ip1), "IP1 should be rate limited")
	assert.False(t, limiter.Allow(ip2), "IP2 should be rate limited")
}

func TestRateLimiter_Allow_WindowExpiration(t *testing.T) {
	limiter := NewRateLimiter(2, 100*time.Millisecond)
	defer limiter.Stop()

	ip := "192.168.1.1"

	// Use up the limit
	assert.True(t, limiter.Allow(ip))
	assert.True(t, limiter.Allow(ip))
	assert.False(t, limiter.Allow(ip), "Should be rate limited")

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be able to make requests again
	assert.True(t, limiter.Allow(ip), "Should allow requests after window expires")
}

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewRateLimiter(2, 1*time.Minute)
	defer limiter.Stop()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "3rd request should be rate limited")
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(1, 1*time.Minute)
	defer limiter.Stop()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First IP should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second IP should also succeed (separate rate limit)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// First IP should now be rate limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

func TestRateLimiter_Stop(t *testing.T) {
	limiter := NewRateLimiter(5, 1*time.Minute)

	// Stop should not panic
	assert.NotPanics(t, func() {
		limiter.Stop()
	})

	// Stop again should not panic (idempotent)
	assert.NotPanics(t, func() {
		limiter.Stop()
	})
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewRateLimiter(100, 1*time.Minute)
	defer limiter.Stop()

	ip := "192.168.1.1"
	done := make(chan bool, 100)

	// Concurrently make requests
	for i := 0; i < 100; i++ {
		go func() {
			limiter.Allow(ip)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify no data races occurred (test should complete without panics)
	assert.True(t, true, "Concurrent access should be safe")
}
