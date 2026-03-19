package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	token     string
	apiServer string
)

var rootCmd = &cobra.Command{
	Use:   "ship",
	Short: "SHIP Platform CLI",
	Long:  `A command line interface for interacting with the SHIP Platform.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Personal Access Token for authentication")
	rootCmd.PersistentFlags().StringVar(&apiServer, "server", "https://api.ship-platform.com", "API Server URL")
}
