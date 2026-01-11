// // internal/api/rest/middlewares/rate_limit.go
package middlewares

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"strconv"
// 	"sync"
// 	"time"

// 	"budget-planner/pkg/logger"

// 	"github.com/gin-gonic/gin"
// 	// "github.com/go-redis/redis/v8"
// )

// // RateLimiter defines the interface for rate limiting implementations
// type RateLimiter interface {
// 	Allow(key string) (bool, int, time.Duration)
// }

// // MemoryRateLimiter implements in-memory rate limiting (for development or small deployments)
// type MemoryRateLimiter struct {
// 	mu          sync.RWMutex
// 	requests    map[string][]time.Time
// 	limit       int
// 	window      time.Duration
// 	cleanupChan chan bool
// }

// // NewMemoryRateLimiter creates a new memory-based rate limiter
// func NewMemoryRateLimiter(limit int, window time.Duration) *MemoryRateLimiter {
// 	limiter := &MemoryRateLimiter{
// 		requests:    make(map[string][]time.Time),
// 		limit:       limit,
// 		window:      window,
// 		cleanupChan: make(chan bool),
// 	}

// 	// Start cleanup goroutine
// 	go limiter.cleanup()

// 	return limiter
// }

// // cleanup periodically removes expired entries
// func (l *MemoryRateLimiter) cleanup() {
// 	ticker := time.NewTicker(time.Minute)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			l.removeExpired()
// 		case <-l.cleanupChan:
// 			return
// 		}
// 	}
// }

// // removeExpired removes timestamps older than the window
// func (l *MemoryRateLimiter) removeExpired() {
// 	cutoff := time.Now().Add(-l.window)
// 	l.mu.Lock()
// 	defer l.mu.Unlock()

// 	for key, times := range l.requests {
// 		var validTimes []time.Time
// 		for _, t := range times {
// 			if t.After(cutoff) {
// 				validTimes = append(validTimes, t)
// 			}
// 		}

// 		if len(validTimes) == 0 {
// 			delete(l.requests, key)
// 		} else {
// 			l.requests[key] = validTimes
// 		}
// 	}
// }

// // Allow checks if a request is allowed based on the rate limit
// func (l *MemoryRateLimiter) Allow(key string) (bool, int, time.Duration) {
// 	now := time.Now()
// 	cutoff := now.Add(-l.window)

// 	l.mu.Lock()
// 	defer l.mu.Unlock()

// 	// Get existing requests for this key
// 	times, exists := l.requests[key]
// 	if !exists {
// 		times = []time.Time{}
// 	}

// 	// Filter out expired timestamps
// 	var validTimes []time.Time
// 	for _, t := range times {
// 		if t.After(cutoff) {
// 			validTimes = append(validTimes, t)
// 		}
// 	}

// 	// Check if we're over the limit
// 	remaining := l.limit - len(validTimes)
// 	if remaining <= 0 {
// 		// Calculate reset time (when the oldest request expires)
// 		resetAfter := validTimes[0].Sub(cutoff)
// 		return false, 0, resetAfter
// 	}

// 	// Allow the request and record timestamp
// 	validTimes = append(validTimes, now)
// 	l.requests[key] = validTimes
// 	return true, remaining - 1, 0
// }

// // RedisRateLimiter implements distributed rate limiting using Redis
// type RedisRateLimiter struct {
// 	redisClient *redis.Client
// 	limit       int
// 	window      time.Duration
// 	prefix      string
// }

// // NewRedisRateLimiter creates a new Redis-based rate limiter
// func NewRedisRateLimiter(redisClient *redis.Client, limit int, window time.Duration, prefix string) *RedisRateLimiter {
// 	return &RedisRateLimiter{
// 		redisClient: redisClient,
// 		limit:       limit,
// 		window:      window,
// 		prefix:      prefix,
// 	}
// }

// // Allow checks if a request is allowed based on the rate limit (Redis implementation)
// func (l *RedisRateLimiter) Allow(key string) (bool, int, time.Duration) {
// 	ctx := context.Background()
// 	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)
// 	now := time.Now().UnixNano()
// 	windowMicro := l.window.Microseconds()

// 	// Remove expired timestamps and add current request
// 	pipeline := l.redisClient.Pipeline()
// 	pipeline.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(now-windowMicro, 10))
// 	pipeline.ZAdd(ctx, redisKey, &redis.Z{Score: float64(now), Member: now})
// 	pipeline.ZCard(ctx, redisKey)
// 	pipeline.Expire(ctx, redisKey, l.window+time.Minute)
// 	results, err := pipeline.Exec(ctx)

// 	if err != nil {
// 		// Fall back to allowing the request on Redis errors
// 		return true, l.limit - 1, 0
// 	}

// 	// Get the count of current requests
// 	count := results[2].(*redis.IntCmd).Val()
// 	remaining := l.limit - int(count)

// 	if remaining < 0 {
// 		// Get the oldest timestamp to calculate reset time
// 		oldestCmd := l.redisClient.ZRange(ctx, redisKey, 0, 0)
// 		oldest, err := oldestCmd.Result()

// 		if err != nil || len(oldest) == 0 {
// 			return false, 0, l.window
// 		}

// 		oldestTime, _ := strconv.ParseInt(oldest[0], 10, 64)
// 		resetAfter := time.Duration(oldestTime + windowMicro - now)
// 		return false, 0, resetAfter
// 	}

// 	return true, remaining, 0
// }

// // RateLimitMiddleware provides rate limiting for API endpoints
// type RateLimitMiddleware struct {
// 	limiter RateLimiter
// 	logger  *logger.Logger
// }

// // NewRateLimitMiddleware creates a new rate limiting middlewares
// func NewRateLimitMiddleware(limiter RateLimiter, logger *logger.Logger) *RateLimitMiddleware {
// 	return &RateLimitMiddleware{
// 		limiter: limiter,
// 		logger:  logger,
// 	}
// }

// // Limit applies rate limiting based on configurable parameters
// func (m *RateLimitMiddleware) Limit() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Get identifier for rate limiting (IP by default, could be user ID, API key, etc.)
// 		identifier := c.ClientIP()

// 		// For authenticated requests, you might want to use user ID instead
// 		if userID, exists := c.Get("userID"); exists {
// 			identifier = userID.(string)
// 		} else if clientID, exists := c.Get("clientID"); exists {
// 			identifier = clientID.(string)
// 		}

// 		// Create unique key for this endpoint
// 		key := fmt.Sprintf("%s:%s", c.FullPath(), identifier)

// 		// Check if request is allowed
// 		allowed, remaining, retryAfter := m.limiter.Allow(key)

// 		// Set rate limit headers
// 		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

// 		if !allowed {
// 			c.Header("X-RateLimit-Reset", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
// 			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))

// 			m.logger.WithFields(logger.Fields{
// 				"ip":         c.ClientIP(),
// 				"path":       c.FullPath(),
// 				"retryAfter": retryAfter.String(),
// 			}).Info("Rate limit exceeded")

// 			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
// 				"error":   "Too Many Requests",
// 				"message": "Rate limit exceeded. Please try again later.",
// 			})
// 			return
// 		}

// 		c.Next()
// 	}
// }

// // PerClientRateLimit allows different rate limits for different clients or APIs
// func (m *RateLimitMiddleware) PerClientRateLimit(limiterFactory func(clientID string) RateLimiter) gin.HandlerFunc {
// 	limiterCache := make(map[string]RateLimiter)
// 	var mu sync.RWMutex

// 	return func(c *gin.Context) {
// 		// Determine client identifier (API key ID, user type, etc.)
// 		var clientID string
// 		if id, exists := c.Get("clientID"); exists {
// 			clientID = id.(string)
// 		} else if userID, exists := c.Get("userID"); exists {
// 			clientID = "user:" + userID.(string)
// 		} else {
// 			clientID = "anonymous"
// 		}

// 		// Get or create limiter for this client
// 		mu.RLock()
// 		limiter, exists := limiterCache[clientID]
// 		mu.RUnlock()

// 		if !exists {
// 			limiter = limiterFactory(clientID)
// 			mu.Lock()
// 			limiterCache[clientID] = limiter
// 			mu.Unlock()
// 		}

// 		// Apply the limiter
// 		key := c.FullPath()
// 		allowed, remaining, retryAfter := limiter.Allow(key)

// 		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

// 		if !allowed {
// 			c.Header("X-RateLimit-Reset", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
// 			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))

// 			m.logger.WithFields(logger.Fields{
// 				"client":     clientID,
// 				"path":       c.FullPath(),
// 				"retryAfter": retryAfter.String(),
// 			}).Info("Client rate limit exceeded")

// 			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
// 				"error":   "Too Many Requests",
// 				"message": "Rate limit exceeded. Please try again later.",
// 			})
// 			return
// 		}

// 		c.Next()
// 	}
// }

