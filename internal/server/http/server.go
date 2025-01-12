package http

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go-winx-api/config"
	"go-winx-api/internal/server/http/middleware"
	"go-winx-api/internal/server/http/routes"
	"go.uber.org/zap"
)

type Server struct {
	App *fiber.App
	Log *zap.Logger
}

func NewServer(log *zap.Logger) *Server {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	app.Use(middleware.RequestLogger(log))

	app.Static("/", "./docs")

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
