// package middleware

// import (
// 	"net/http"
// 	"strconv"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/imraushankr/brevity/server/src/configs"
// 	"github.com/ulule/limiter/v3"
// 	"github.com/ulule/limiter/v3/drivers/store/memory"
// )

// // RateLimiterMiddleware creates a configurable rate limiter
// func RateLimiterMiddleware(cfg *configs.RateLimitConfig) gin.HandlerFunc {
// 	if !cfg.Enabled {
// 		return func(c *gin.Context) {
// 			c.Next()
// 		}
// 	}

// 	// Parse rate limit window
// 	window, err := time.ParseDuration(cfg.Window)
// 	if err != nil {
// 		window = time.Minute // Default to 1 minute
// 	}

// 	rate := limiter.Rate{
// 		Period: window,
// 		Limit:  int64(cfg.Requests),
// 	}
// 	store := memory.NewStore()
// 	limiterInstance := limiter.New(store, rate)

// 	return func(c *gin.Context) {
// 		context, err := limiterInstance.Get(c, c.ClientIP())
// 		if err != nil {
// 			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
// 				"error": "Internal server error",
// 			})
// 			return
// 		}

// 		// Set rate limit headers
// 		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
// 		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
// 		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

// 		if context.Reached {
// 			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
// 				"error": "Too many requests",
// 			})
// 			return
// 		}

// 		c.Next()
// 	}
// }


package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	limiter := rate.NewLimiter(r, b)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "Too many requests",
			})
			return
		}
		c.Next()
	}
}