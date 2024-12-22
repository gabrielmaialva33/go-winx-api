package main

import (
	"fmt"
	"go-winx-api/config"
	"os"

	"github.com/spf13/cobra"
	"go-winx-api/cmd"
	"go-winx-api/internal/utils"
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
