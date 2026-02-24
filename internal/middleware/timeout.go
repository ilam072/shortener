package middleware

import (
	"context"
	"errors"
	"github.com/wb-go/wbf/ginext"
	"net/http"
	"time"
)

func TimeoutMiddleware(timeout time.Duration) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, ginext.H{
				"error": "request timed out",
			})
			return
		}
	}
}
