package redis

import (
	"context"
	"fmt"
	"net/http"
	"time"

	rdc "github.com/redis/go-redis/v9"
)

// RateLimiter implements sliding window rate limiting using Redis
type RateLimiter struct {
	rdb        *rdc.Client
	maxReqs    int
	window     time.Duration
}

// NewRateLimiter creates a rate limiter: 100 requests per minute per user
func NewRateLimiter(rdb *rdc.Client) *RateLimiter {
	return &RateLimiter{
		rdb:     rdb,
		maxReqs: 100,
		window:  1 * time.Minute,
	}
}

// Middleware provides HTTP rate limiting using Redis sorted sets (sliding window)
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := getUserID(ctx)
		if userID == "" {
			userID = r.RemoteAddr // fallback to IP
		}

		key := fmt.Sprintf("ratelimit:%s", userID)
		now := time.Now().UnixMicro()
		windowStart := float64(time.Now().Add(-rl.window).UnixMicro())

		pipe := rl.rdb.Pipeline()
		// Remove expired entries
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%f", windowStart))
		// Add current request
		pipe.ZAdd(ctx, key, rdc.Z{Score: float64(now), Member: now})
		// Count requests in window
		countCmd := pipe.ZCard(ctx, key)
		// Set expiry on key
		pipe.Expire(ctx, key, rl.window)

		_, err := pipe.Exec(ctx)
		if err != nil {
			// Redis error — allow request through (fail open)
			next.ServeHTTP(w, r)
			return
		}

		count := countCmd.Val()
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxReqs))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(rl.maxReqs)-count)))

		if count > int64(rl.maxReqs) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded, try again later"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getUserID(ctx context.Context) string {
	if uid, ok := ctx.Value("user_id").(string); ok {
		return uid
	}
	return ""
}
