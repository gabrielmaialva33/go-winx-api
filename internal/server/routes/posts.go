package routes

import (
	"github.com/gofiber/fiber/v2"
	"go-winx-api/internal/server/handlers"
	"go-winx-api/internal/telegram"
	"go.uber.org/zap"
)

func registerPostRoutes(app *fiber.App, log *zap.Logger, repository *telegram.Repository) {

	api := app.Group("/api/v1")

	api.Get("/posts", handlers.GetAllPosts(log, repository))
}
