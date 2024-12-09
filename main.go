package main

import (
	"github.com/spf13/cobra"
	"go-winx-api/cmd"
)

var rootCmd *cobra.Command

func init() {

	rootCmd = cmd.NewRootCommand()

}

func main() {

	_ = rootCmd.Execute()

}
