package cmd

import "github.com/spf13/cobra"

const version = "0.0.1"

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:               "winx",
		Short:             "CineWing File Stream Service",
		Long:              `CineWing is a file stream service for movies and series. It's based on Telegram API.`,
		Version:           version,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				return
			}
		},
	}

	return root
}
