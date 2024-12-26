package handlers

import (
	"context"
	"go-winx-api/internal/telegram"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func GetAllPosts(log *zap.Logger, repository *telegram.Repository) fiber.Handler {
	ctx := context.Background()
	repository.GetHistory(ctx)

	return func(c *fiber.Ctx) error {
		log.Info("Get all posts")
		return c.JSON(fiber.Map{"message": "Get all posts"})
	}
}
