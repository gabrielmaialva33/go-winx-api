package handlers

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go-winx-api/internal/models"
	"go-winx-api/internal/services/telegram"
	"go.uber.org/zap"
	"strconv"
	"strings"
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'message_id' parameter",
			})
		}

		log.Info("Streaming video", zap.Int("message_id", messageID))

		file, err := repository.GetFile(context.Background(), messageID)
		if err != nil {
			log.Error("Failed to fetch file metadata", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch file metadata",
			})
		}

		rangeHeader := c.Get("Range")
		var start, end int64
		if rangeHeader != "" {
			parts := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
			start, _ = strconv.ParseInt(parts[0], 10, 64)
			if len(parts) > 1 && parts[1] != "" {
				end, _ = strconv.ParseInt(parts[1], 10, 64)
			} else {
				end = file.FileSize - 1
			}
		} else {
			start = 0
			end = file.FileSize - 1
		}

		if end >= file.FileSize {
			end = file.FileSize - 1
		}

		chunkSize := end - start + 1

		c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.FileSize))
		c.Set("Content-Length", fmt.Sprintf("%d", chunkSize))
		c.Set("Accept-Ranges", "bytes")
		c.Set("Content-Type", file.MimeType)
		c.Status(fiber.StatusPartialContent)

		stream, err := repository.GetVideoStream(context.Background(), file, start, end)
		if err != nil {
			log.Error("Failed to stream video", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to stream video",
			})
		}

		return c.SendStream(stream)
	}
}
