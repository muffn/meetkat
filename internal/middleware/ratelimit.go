package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens    float64
	lastCheck time.Time
}

// RateLimiter is a per-IP token bucket rate limiter backed by stdlib only.
type RateLimiter struct {
	mu      sync.Mutex
	buckets sync.Map
	rate    float64 // tokens per second
	burst   float64 // max burst size
}

// NewRateLimiter creates a limiter that allows ratePerMinute requests per minute
// with an initial burst capacity of burst tokens.
func NewRateLimiter(ratePerMinute, burst int) *RateLimiter {
	return &RateLimiter{
		rate:  float64(ratePerMinute) / 60.0,
		burst: float64(burst),
	}
}

// Middleware returns a Gin handler that enforces the rate limit per client IP.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		val, _ := rl.buckets.LoadOrStore(ip, &bucket{tokens: rl.burst, lastCheck: now})
		b := val.(*bucket)

		rl.mu.Lock()
		elapsed := now.Sub(b.lastCheck).Seconds()
		b.tokens = minFloat(rl.burst, b.tokens+elapsed*rl.rate)
		b.lastCheck = now
		if b.tokens < 1 {
			rl.mu.Unlock()
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		b.tokens--
		rl.mu.Unlock()

		c.Next()
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
