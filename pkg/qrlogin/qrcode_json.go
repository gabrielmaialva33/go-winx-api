package qrlogin

import (
	"context"

	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
)

type QRInfo struct {
	QRCodeURL string `json:"qr_code_url"`
	ExpiresIn int64  `json:"expires_in"`
	Session   string `json:"session,omitempty"`
	Error     string `json:"error,omitempty"`
}

func GenerateQRSessionJSON(apiId int, apiHash string) (*QRInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dispatcher := tg.NewUpdateDispatcher()
	client := telegram.NewClient(apiId, apiHash, telegram.Options{UpdateHandler: dispatcher})
	sessionStorage := &session.StorageMemory{}
	qrInfo := &QRInfo{}

	err := client.Run(ctx, func(ctx context.Context) error {
		_, err := client.QR().Auth(ctx, qrlogin.OnLoginToken(dispatcher), func(ctx context.Context, token qrlogin.Token) error {
			qrInfo.QRCodeURL = token.URL()
			qrInfo.ExpiresIn = int64(time.Until(token.Expires()).Seconds())
			return nil
		})
		if err != nil {
			qrInfo.Error = err.Error()
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	sessionData, _ := sessionStorage.LoadSession(ctx)
	if sessionData != nil {
		qrInfo.Session = string(sessionData)
	}

	return qrInfo, nil
}
