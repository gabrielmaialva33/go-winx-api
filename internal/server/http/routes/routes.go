package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func SetupRoutes(app *fiber.App, log *zap.Logger) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs")
	})

	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/redoc.html")
	})

	registerPostRoutes(app, log)
}
