package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hamfa/task-manager/internal/model"
)

// Recovery returns a gin middleware that recovers from panics
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
					Error:   "internal_error",
					Message: "An unexpected error occurred",
					Code:    http.StatusInternalServerError,
				})
			}
		}()

		c.Next()
	}
}
