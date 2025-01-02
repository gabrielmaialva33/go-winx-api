package routes

import (
	"github.com/gofiber/fiber/v2"
	"go-winx-api/internal/server/http/handlers"
	"go-winx-api/internal/services/telegram"
	"go.uber.org/zap"
)

func registerPostRoutes(app *fiber.App, log *zap.Logger) {

	api := app.Group("/api/v1")

	repository := telegram.NewRepository(log)

	api.Get("/posts", handlers.GetAllPosts(log, repository))
	api.Get("/posts/:message_id", handlers.GetPost(log, repository))
	api.Get("/posts/images/:message_id", handlers.GetPostImage(log, repository))
	api.Get("/posts/videos/:message_id", handlers.GetPostVideo(log, repository))
}
