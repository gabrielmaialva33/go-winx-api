package main

import (
	"fmt"
	"os"

	"go-winx-api/cmd"
	"go-winx-api/config"
	"go-winx-api/internal/utils"

	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func init() {
	utils.InitLogger()
	log := utils.Logger

	rootCmd = cmd.NewRootCommand()
	config.Load(log, rootCmd)

}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
