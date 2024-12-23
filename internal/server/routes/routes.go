package routes

import (
	"github.com/gofiber/fiber/v2"
	"go-winx-api/config"
	"go-winx-api/pkg/qrlogin"
	"go.uber.org/zap"
)

func SetupRoutes(app *fiber.App, log *zap.Logger) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, World!",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Post("/login/qrcode", func(c *fiber.Ctx) error {
		apiID := int(config.ValueOf.ApiID)
		apiHash := config.ValueOf.ApiHash

		qrURL, err := qrlogin.GenerateQRSessionURL(apiID, apiHash, log)
		if err != nil {
			log.Error("error generating QR Code", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to generate QR Code",
			})
		}

		return c.JSON(fiber.Map{
			"qr_url": qrURL,
		})
	})
}
