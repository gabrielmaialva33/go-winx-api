package cmd

import (
	"github.com/spf13/cobra"
)

func NewStringSessionCommand() *cobra.Command {
	var appID int
	var appHash string
	var phone string

	session := &cobra.Command{
		Use:   "session",
		Short: "Generate a string session.",
		Long:  "Generate a string session for your Telegram account. Use this session string to authenticate your bot.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	session.Flags().StringVarP(&phone, "phone", "p", "", "Número de telefone com código do país (ex: +5511999999999)")
	session.Flags().IntVar(&appID, "app_id", 0, "APP_ID do Telegram (obtido em https://my.telegram.org)")
	session.Flags().StringVar(&appHash, "app_hash", "", "APP_HASH do Telegram (obtido em https://my.telegram.org)")

	return session
}
