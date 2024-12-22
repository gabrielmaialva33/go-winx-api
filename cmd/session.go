package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go-winx-api/config"
	"go-winx-api/internal/pkg/qrlogin"
)

func NewStringSessionCommand() *cobra.Command {
	var phone string

	session := &cobra.Command{
		Use:   "session",
		Short: "Generate a string session.",
		Long:  "Generate a string session for your Telegram account. Use this session string to authenticate your bot.",
		RunE: func(cmd *cobra.Command, args []string) error {

			appID := int(config.ValueOf.ApiID)
			appHash := config.ValueOf.ApiHash

			err := qrlogin.GenerateQRSession(appID, appHash)
			if err != nil {
				return fmt.Errorf("failed to generate session via phone login: %w", err)
			}

			return nil
		},
	}

	session.Flags().StringVar(&phone, "phone", "", "Phone number with country code (ex: +5511999999999)")

	return session
}
