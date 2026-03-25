package cmd

import (
	"fmt"
	"os"

	"ship-cli/config"

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			cfg = &config.Config{}
		}

		serverFlag := cmd.PersistentFlags().Lookup("server")
		if !serverFlag.Changed && cfg.Server != "" {
			apiServer = cfg.Server
		}

		// If token is not provided via flag, try to load it from config
		if token == "" {
			if cfg.Token != "" {
				token = cfg.Token
			}
		} else if cfg.Token != token {
			cfg.Token = token
			_ = config.SaveConfig(cfg)
		}

		if serverFlag.Changed && cfg.Server != apiServer {
			cfg.Server = apiServer
			_ = config.SaveConfig(cfg)
		}
	},
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
