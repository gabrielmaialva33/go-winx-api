package http

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go-winx-api/config"
	"go-winx-api/internal/server/http/middleware"
	"go-winx-api/internal/server/http/routes"
	"go-winx-api/internal/services/telegram"
	"go.uber.org/zap"
)

type Server struct {
	App        *fiber.App
	Log        *zap.Logger
	Repository *telegram.Repository
}

func NewServer(log *zap.Logger, repository *telegram.Repository) *Server {
	app := fiber.New()

	app.Use(middleware.RequestLogger(log))

	app.Static("/", "./docs")

	routes.SetupRoutes(app, log, repository)

	return &Server{
		App:        app,
		Log:        log,
		Repository: repository,
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
