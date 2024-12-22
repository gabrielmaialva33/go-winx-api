package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go-winx-api/config"
	"go-winx-api/internal/pkg/qrlogin"
	"go-winx-api/internal/terminal"
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

			if appID == 0 || appHash == "" {
				return errors.New("missing API_ID or API_HASH configuration. Set them in the environment variables or configuration file")
			}

			if phone == "" {
				fmt.Println("Phone number not provided. You can proceed using QR Login or provide a phone number.")

				useQr := terminal.InputPrompt("Do you want to proceed with QR login? (yes/no)")
				if useQr == "yes" {

					err := qrlogin.GenerateQRSession(appID, appHash)
					if err != nil {
						return fmt.Errorf("failed to generate session via QR Login: %w", err)
					}
					return nil
				} else {
					return errors.New("session generation aborted. Please provide a phone number or use QR Login")
				}
			}

			fmt.Println("Starting session generation with phone login...")
			err := qrlogin.GenerateQRSession(appID, appHash)
			if err != nil {
				return fmt.Errorf("failed to generate session via phone login: %w", err)
			}

			fmt.Println("Session generated successfully!")
			return nil
		},
	}

	session.Flags().StringVar(&phone, "phone", "", "Phone number with country code (ex: +5511999999999)")

	return session
}
