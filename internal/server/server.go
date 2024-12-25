package server

import (
	"fmt"
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

	app.Static("/", "./static")

	routes.SetupRoutes(app, log)

	return &Server{
		App: app,
		Log: log,
	}
}

func (s *Server) Start() {
	port := fmt.Sprintf(":%d", config.ValueOf.Port)
	log := s.Log.Named("server")
	log.Sugar().Infof("server is running at %s", port)
	if err := s.App.Listen(port); err != nil {
		log.Fatal("error while starting server", zap.Error(err))
	}
}
