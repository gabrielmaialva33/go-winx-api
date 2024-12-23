package server

import (
	"fmt"

	"log"

	"go-winx-api/config"
	"go-winx-api/internal/server/middleware"
	"go-winx-api/internal/server/routes"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Server struct {
	App *fiber.App
	Log *zap.Logger
}

func NewServer(log *zap.Logger) *Server {
	app := fiber.New()

	app.Use(middleware.RequestLogger(log))

	routes.SetupRoutes(app, log)

	return &Server{
		App: app,
		Log: log,
	}
}

func (s *Server) Start() {
	port := fmt.Sprintf(":%d", config.ValueOf.Port)
	s.Log.Sugar().Infof("Server is running at %s", port)
	if err := s.App.Listen(port); err != nil {
		log.Fatalf("fiber app failed to start: %v", err)
	}
}
