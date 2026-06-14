package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RequestLogger logs one structured line per request after it completes,
// including method, path, final status code, and how long it took.
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next() // run the rest of the chain (and the handler)

		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", time.Since(start)),
			zap.String("request_id", GetRequestID(c)),
		}
		if err != nil {
			log.Error("request failed", append(fields, zap.Error(err))...)
		} else {
			log.Info("request completed", fields...)
		}
		return err
	}
}
