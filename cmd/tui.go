package cmd

import (
	"fmt"
	"os"

	"ship-cli/api"
	"ship-cli/config"
	"ship-cli/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI",
	Long:  `Launch the interactive Text User Interface (TUI) for SHIP Platform.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If token is not provided via flag, try to load it from config
		if token == "" {
			cfg, err := config.LoadConfig()
			if err == nil && cfg.Token != "" {
				token = cfg.Token
			}
		}

		client := api.NewClient(apiServer, token)
		m := ui.NewModel(client)

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running program: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
