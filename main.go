package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"go-winx-api/cmd"
	"os"
)

var rootCmd *cobra.Command

func init() {

	rootCmd = cmd.NewRootCommand()

}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
