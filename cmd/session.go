package cmd

import "github.com/spf13/cobra"

func NewStringSessionCommand() *cobra.Command {
	session := &cobra.Command{
		Use:   "session",
		Short: "Generate a string session.",
		Long: "Generate a string session for your telegram account. " +
			"Use this session string to login to your telegram account.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				return
			}
		},
	}

	return session
}
