package routes

import (
	"github.com/gofiber/fiber/v2"

	"ainyx-user-api/internal/handler"
)

// Register wires every route to its handler.
func Register(app *fiber.App, h *handler.UserHandler) {
	app.Get("/health", healthCheck)

	users := app.Group("/users")
	users.Post("/", h.CreateUser)
	users.Get("/:id", h.GetUser)
	// Day 3 will add:
	//   users.Put("/:id", h.UpdateUser)
	//   users.Delete("/:id", h.DeleteUser)
	//   users.Get("/", h.ListUsers)
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
