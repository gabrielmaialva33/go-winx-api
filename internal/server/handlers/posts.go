package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func GetAllPosts(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Info("Get all posts")
		return c.SendString("Get all posts")
	}
}
