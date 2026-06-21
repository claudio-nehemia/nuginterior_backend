package middleware

import (
	"fmt"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"github.com/gin-gonic/gin"
)

// RateLimiter creates a rate limiting middleware using Redis.
// maxRequests is the max number of requests allowed per window.
// window is the sliding window duration.
func RateLimiter(cacheStore cache.Store, maxRequests int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.FullPath()
		key := fmt.Sprintf("%s%s:%s", constants.KeyRateLimit, ip, endpoint)

		count, err := cacheStore.Incr(c.Request.Context(), key)
		if err != nil {
			// If Redis is down, allow the request
			c.Next()
			return
		}

		if count == 1 {
			_ = cacheStore.Expire(c.Request.Context(), key, window)
		}

		if count > maxRequests {
			helper.TooManyRequests(c, "Terlalu banyak request, coba lagi nanti")
			c.Abort()
			return
		}

		c.Next()
	}
}
