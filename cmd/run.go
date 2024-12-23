package cmd

import (
	"github.com/spf13/cobra"
	"go-winx-api/config"
	"go-winx-api/internal/bot"
	"go-winx-api/internal/utils"
	"go.uber.org/zap"
)

func NewRunCommand() *cobra.Command {
	run := &cobra.Command{
		Use:   "run",
		Short: "Run the bot.",
		Long:  `Run the bot and start listening for incoming messages.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.InitLogger()
			log := utils.Logger

			mainLogger := log.Named("Main")
			mainLogger.Info("Starting server")

			config.Load(log, cmd)

			mainBot, err := bot.StartClient(log)
			if err != nil {
				log.Panic("Failed to start main bot", zap.Error(err))
			}

			workers, err := bot.StartWorkers(log)
			if err != nil {
				log.Panic("Failed to start workers", zap.Error(err))
				return err
			}

			workers.AddDefaultClient(mainBot, mainBot.Self)
			bot.StartUserBot(log)

			mainLogger.Info("Server started", zap.Int("port", config.ValueOf.Port))
			mainLogger.Info("File Stream Bot", zap.String("version", versionString))
			mainLogger.Sugar().Infof("Server is running at %s", config.ValueOf.Host)

			if err != nil {
				mainLogger.Sugar().Fatalln(err)
			}

			return nil
		},
	}

	return run

}
