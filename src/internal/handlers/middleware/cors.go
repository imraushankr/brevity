// package middleware

// import (
// 	"time"

// 	"github.com/gin-contrib/cors"
// 	"github.com/gin-gonic/gin"
// 	"github.com/imraushankr/brevity/server/src/configs"
// )

// // CorsMiddleware creates a configurable CORS middleware
// func CorsMiddleware(cfg *configs.CORSConfig) gin.HandlerFunc {
// 	if !cfg.Enabled {
// 		return func(c *gin.Context) {
// 			c.Next()
// 		}
// 	}

// 	// Parse max age duration
// 	maxAge, err := time.ParseDuration(cfg.MaxAge)
// 	if err != nil {
// 		maxAge = 12 * time.Hour // Default fallback
// 	}

// 	return cors.New(cors.Config{
// 		AllowOrigins:     cfg.AllowOrigins,
// 		AllowMethods:     cfg.AllowMethods,
// 		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
// 		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
// 		AllowCredentials: true,
// 		MaxAge:           maxAge,
// 	})
// }

package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}