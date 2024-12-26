package handlers

import (
	"context"
	"go-winx-api/internal/models"
	"go-winx-api/internal/services/telegram"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
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
