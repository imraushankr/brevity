// // /middleware/logger.go
// package middleware

// import (
// 	"time"

// 	"github.com/gin-contrib/requestid"
// 	"github.com/gin-gonic/gin"
// 	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
// )

// // LoggerMiddleware creates a configurable logging middleware
// func LoggerMiddleware(log logger.Logger) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		start := time.Now()
// 		path := c.Request.URL.Path
// 		query := c.Request.URL.RawQuery
// 		method := c.Request.Method
// 		ip := c.ClientIP()
// 		userAgent := c.Request.UserAgent()
// 		reqID := requestid.Get(c)

// 		// Process request
// 		c.Next()

// 		// Collect metrics
// 		latency := time.Since(start)
// 		status := c.Writer.Status()

// 		fields := []logger.Field{
// 			logger.Int("status", status),
// 			logger.String("method", method),
// 			logger.String("path", path),
// 			logger.String("query", query),
// 			logger.String("ip", ip),
// 			logger.String("user-agent", userAgent),
// 			logger.Duration("latency", latency),
// 			logger.String("request-id", reqID),
// 		}

// 		if len(c.Errors) > 0 {
// 			fields = append(fields, logger.Any("errors", c.Errors))
// 		}

// 		switch {
// 		case status >= 500:
// 			log.Error("server error", fields...)
// 		case status >= 400:
// 			log.Warn("client error", fields...)
// 		default:
// 			log.Info("request", fields...)
// 		}
// 	}
// }

package middleware

import (
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
)

func RequestLogger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		
		c.Next()
		
		end := time.Now()
		latency := end.Sub(start)
		
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				log.Error(e)
			}
		} else {
			log.Info(path,
				logger.Int("status", c.Writer.Status()),
				logger.String("method", c.Request.Method),
				logger.String("path", path),
				logger.String("query", query),
				logger.String("ip", c.ClientIP()),
				logger.String("user-agent", c.Request.UserAgent()),
				logger.Duration("latency", latency),
			)
		}
	}
}