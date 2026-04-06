package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type client struct {
	count    int
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*client
	max      int
	window   time.Duration
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		max:     max,
		window:  window,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > rl.window {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		cl, exists := rl.clients[ip]
		if !exists {
			rl.clients[ip] = &client{count: 1, lastSeen: time.Now()}
			rl.mu.Unlock()
			c.Next()
			return
		}

		if time.Since(cl.lastSeen) > rl.window {
			cl.count = 1
			cl.lastSeen = time.Now()
			rl.mu.Unlock()
			c.Next()
			return
		}

		cl.count++
		cl.lastSeen = time.Now()

		if cl.count > rl.max {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, try again later"})
			c.Abort()
			return
		}

		rl.mu.Unlock()
		c.Next()
	}
}
