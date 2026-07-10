package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type rateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration

	stopCh chan struct{}
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
		stopCh: make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

func (r *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.purgeStale()
		case <-r.stopCh:
			return
		}
	}
}

func (r *rateLimiter) purgeStale() {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := time.Now().Add(-2 * r.window)
	for key, hits := range r.hits {
		last := hits[len(hits)-1]
		if last.Before(cutoff) {
			delete(r.hits, key)
		}
	}
}

func (r *rateLimiter) stop() {
	close(r.stopCh)
}

func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-r.window)
	hits := r.hits[key]
	out := hits[:0]
	for _, t := range hits {
		if t.After(cutoff) {
			out = append(out, t)
		}
	}
	if len(out) >= r.limit {
		r.hits[key] = out
		return false
	}
	out = append(out, now)
	r.hits[key] = out
	return true
}

var loginLimiter = newRateLimiter(10, time.Minute)

func LoginRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !loginLimiter.allow(c.IP()) {
			return c.Status(429).JSON(fiber.Map{"error": "rate_limited", "message": "terlalu banyak percobaan, coba lagi nanti"})
		}
		return c.Next()
	}
}

type tokenRateLimiterStore struct {
	mu       sync.Mutex
	limiters map[int]*rateLimiter
}

var globalTokenLimiters = &tokenRateLimiterStore{
	limiters: make(map[int]*rateLimiter),
}

func TokenRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate, _ := c.Locals("token_rate").(int)
		if rate <= 0 {
			rate = 60
		}

		globalTokenLimiters.mu.Lock()
		rl, ok := globalTokenLimiters.limiters[rate]
		if !ok {
			rl = newRateLimiter(rate, time.Minute)
			globalTokenLimiters.limiters[rate] = rl
		}
		globalTokenLimiters.mu.Unlock()

		tokID, _ := c.Locals("token_id").(string)
		if tokID == "" {
			tokID = UserID(c)
		}
		if !rl.allow("token:"+tokID) {
			return c.Status(429).JSON(fiber.Map{"error": "rate_limited", "message": "rate limit token tercapai"})
		}
		return c.Next()
	}
}
