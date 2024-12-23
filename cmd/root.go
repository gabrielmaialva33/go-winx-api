package cmd

import "github.com/spf13/cobra"

const versionString = "0.0.1"

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:               "winx",
		Short:             "CineWing File Stream Service",
		Long:              `CineWing is a file stream service for movies and series. It's based on Telegram API.`,
		Version:           versionString,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				return
			}
		},
	}

	root.AddCommand(NewStringSessionCommand())
	root.AddCommand(NewRunCommand())

	return root
}
