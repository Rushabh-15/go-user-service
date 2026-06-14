package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the response header carrying the correlation id.
	RequestIDHeader = "X-Request-Id"
	// requestIDKey is the Locals key under which the id is stored.
	requestIDKey = "request_id"
)

// RequestID assigns a unique id to each request, stores it for handlers and
// logging to read, and echoes it back in the response header.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := uuid.NewString()
		c.Locals(requestIDKey, id)
		c.Set(RequestIDHeader, id)
		return c.Next()
	}
}

// GetRequestID returns the id stored by RequestID, or "" if absent.
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals(requestIDKey).(string); ok {
		return id
	}
	return ""
}
