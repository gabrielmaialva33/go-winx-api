package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func GetAllPosts(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Info("Fetching all posts")
		return c.JSON(fiber.Map{
			"message": "Get all posts",
		})
	}
}
