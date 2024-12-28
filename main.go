package main

import (
	"go-winx-api/config"
	"go-winx-api/internal/cache"
	"go-winx-api/internal/server/http"
	"go-winx-api/internal/services/telegram"
	"go-winx-api/internal/utils"

	"go.uber.org/zap"
)

func main() {

	utils.InitLogger()
	log := utils.Logger

	logger := log.Named("main")
	logger.Info("starting server")

	config.Load(log)

	client, err := telegram.InitClient(log)
	if err != nil {
		logger.Fatal("error while starting telegram client", zap.Error(err))
	}

	telegram.TgClient = client

	cache.InitCache(log)

	workers, err := telegram.StartWorkers(log)
	if err != nil {
		log.Panic("failed to start workers", zap.Error(err))
		return
	}

	workers.AddDefaultClient(client, client.Self)

	logger.Info("server started", zap.Int("port", config.ValueOf.Port))
	logger.Sugar().Infof("server is running at %s", config.ValueOf.Host)

	repository := telegram.NewRepository(client, log)

	s := http.NewServer(log, repository)
	s.Start()

}
