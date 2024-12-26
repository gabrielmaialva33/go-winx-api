package routes

import (
	"github.com/gofiber/fiber/v2"
	"go-winx-api/internal/services/telegram"
	"go.uber.org/zap"
)

func SetupRoutes(app *fiber.App, log *zap.Logger, repository *telegram.Repository) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs")
	})

	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/redoc.html")
	})

	registerPostRoutes(app, log, repository)
}
