package handlers

import (
	"context"
	"fmt"
	"go-winx-api/internal/models"
	"go-winx-api/internal/services/telegram"
	"net/http"
	"strconv"
	"strings"

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

		messageID, err := strconv.Atoi(c.Params("message_id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid 'message_id' parameter"})
		}

		size, err := strconv.Atoi(c.Params("size"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid 'size' parameter"})
		}

		rangeHeader := c.Get("Range", "")
		start, end := 0, size-1

		if rangeHeader != "" {
			if !strings.HasPrefix(rangeHeader, "bytes=") {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid 'Range' header"})
			}

			rangeParts := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
			start, _ = strconv.Atoi(rangeParts[0])
			if len(rangeParts) > 1 && rangeParts[1] != "" {
				end, _ = strconv.Atoi(rangeParts[1])
			}
			if end >= size {
				end = size - 1
			}
		}

		if start >= size || start > end {
			return c.Status(http.StatusRequestedRangeNotSatisfiable).JSON(fiber.Map{"error": "Invalid range"})
		}

		chunkSize := end - start + 1
		log.Info("Streaming video", zap.Int("message_id", messageID), zap.Int("start", start), zap.Int("end", end), zap.Int("chunk_size", chunkSize))

		ctx := context.Background()

		c.Set("Content-Type", "video/mp4")
		c.Set("Accept-Ranges", "bytes")
		c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
		c.Set("Content-Length", strconv.Itoa(chunkSize))
		c.Status(http.StatusPartialContent)

		if err := repository.StreamVideo(ctx, messageID, c.Response().BodyWriter(), start, end); err != nil {
			log.Error("Failed to stream video", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to stream video"})
		}

		return nil
	}
}
