package handlers

import (
	"context"
	"go-winx-api/internal/services/telegram"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func GetAllPosts(log *zap.Logger, repository *telegram.Repository) fiber.Handler {
	log = log.Named("posts")
	log.Sugar().Info("registering posts routes")

	ctx := context.Background()
	messages, err := repository.GetHistory(ctx)
	if err != nil {
		return nil
	}

	return func(c *fiber.Ctx) error {
		return c.JSON(messages)
	}
}
