package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func SetupRoutes(app *fiber.App, log *zap.Logger) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, World!",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})
}
