package telegram

import (
	"context"
	"time"

	"go-winx-api/config"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
)

var TgClient *gotgproto.Client

func InitClient() (*gotgproto.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	clientChan := make(chan struct {
		client *gotgproto.Client
		err    error
	})

	session := sessionMaker.TelethonSession(config.ValueOf.UserSession).Name("main.session")
	go func(ctx context.Context) {
		client, err := gotgproto.NewClient(
			config.ValueOf.ApiId,
			config.ValueOf.ApiHash,
			gotgproto.ClientTypePhone(""),
			&gotgproto.ClientOpts{
				Session:          session,
				DisableCopyright: true,
			},
		)
		clientChan <- struct {
			client *gotgproto.Client
			err    error
		}{client, err}
	}(ctx)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-clientChan:
		if result.err != nil {
			return nil, result.err
		}

		log.Info("client started", zap.String("username", result.client.Self.Username))

		TgClient = result.client
		return result.client, nil
	}
}
