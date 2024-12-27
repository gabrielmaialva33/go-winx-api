package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go-winx-api/internal/models"
	"go-winx-api/internal/services/telegram"
	"go.uber.org/zap"
	"strconv"
)

func GetAllPosts(log *zap.Logger, repository *telegram.Repository) fiber.Handler {
	log = log.Named("posts")

	return func(c *fiber.Ctx) error {
		perPage, err := strconv.Atoi(c.Query("per_page", "10"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'per_page' parameter",
			})
		}

		offsetId, err := strconv.Atoi(c.Query("offset_id", "0"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'offset_id' parameter",
			})
		}

		log.Info("Fetching posts", zap.Int("per_page", perPage), zap.Int("offset_id", offsetId))

		ctx := context.Background()

		pagination := models.PaginationData{
			PerPage:  perPage,
			OffsetId: offsetId,
		}

		messages, err := repository.PaginatePosts(ctx, pagination)
		if err != nil {
			log.Error("Failed to fetch posts", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch posts",
			})
		}

		return c.JSON(messages)
	}
}

func GetPost(log *zap.Logger, repository *telegram.Repository) fiber.Handler {
	log = log.Named("post")

	return func(c *fiber.Ctx) error {
		messageId, err := strconv.Atoi(c.Params("message_id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'id' parameter",
			})
		}

		log.Info("Fetching post", zap.Int("id", messageId))

		ctx := context.Background()

		message, err := repository.GetPost(ctx, messageId)
		if err != nil {
			log.Error("failed to fetch post", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch post",
			})
		}

		return c.JSON(message)
	}
}

func GetPostImage(log *zap.Logger, repository *telegram.Repository) fiber.Handler {
	log = log.Named("stream_images")

	return func(c *fiber.Ctx) error {
		messageID, err := strconv.Atoi(c.Params("message_id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'message_id' parameter",
			})
		}

		log.Info("Streaming image", zap.Int("message_id", messageID))

		ctx := context.Background()

		c.Set("Content-Type", "image/jpeg")
		c.Set("Cache-Control", "no-cache")

		if err := repository.GetPostImage(ctx, messageID, c.Response().BodyWriter()); err != nil {
			log.Error("failed to stream image", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to stream image",
			})
		}

		return nil
	}
}

func GetPostVideo(log *zap.Logger, repository *telegram.Repository) fiber.Handler {
	log = log.Named("stream_videos")

	return func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	}
}
