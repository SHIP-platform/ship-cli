package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"ship-cli/api"
	"ship-cli/ui"

	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsTail   int
	logsTui    bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [APP_ID]",
	Short: "Stream logs from a deployed application",
	Long: `Stream logs from the main container of an app running on the SHIP Platform.

Uses GET /api/applications/{id}/logs/stream (Server-Sent Events). Use --follow=false to
fetch available lines and exit. Use --tail=N to limit to the last N lines.

Use --tui for a full-screen scrollable viewer (arrow keys, PgUp/PgDn, mouse wheel).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appID := args[0]

		if token == "" {
			log.Fatal("Error: --token is required (or set token in ~/.ship/config.json)")
		}

		if logsTui {
			client := api.NewClient(strings.TrimSuffix(apiServer, "/"), token)
			if err := ui.RunLogsTUI(client, appID, logsFollow, logsTail); err != nil {
				fmt.Fprintf(os.Stderr, "logs TUI: %v\n", err)
				os.Exit(1)
			}
			return
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println("\nStopped log stream.")
			os.Exit(0)
		}()

		streamLogs(appID)
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVar(&logsFollow, "follow", true, "Keep streaming new log lines until interrupted")
	logsCmd.Flags().IntVar(&logsTail, "tail", 0, "Only include the last N lines (0 = server default / full history)")
	logsCmd.Flags().BoolVarP(&logsTui, "tui", "t", false, "Open logs in a full-screen scrollable TUI")
}

func streamLogs(appID string) {
	base := strings.TrimSuffix(apiServer, "/")
	u, err := url.Parse(base + "/api/applications/" + url.PathEscape(appID) + "/logs/stream")
	if err != nil {
		log.Fatalf("Invalid log URL: %v", err)
	}
	q := u.Query()
	if !logsFollow {
		q.Set("follow", "false")
	}
	if logsTail > 0 {
		q.Set("tail", strconv.Itoa(logsTail))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to connect to log stream: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Error streaming logs: HTTP %d %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nLog stream closed by server.")
				return
			}
			log.Fatalf("Error reading log stream: %v", err)
		}

		if strings.HasPrefix(line, "data: ") {
			content := strings.TrimPrefix(line, "data: ")
			if strings.TrimSpace(content) == "stream ended" {
				return
			}
			fmt.Print(content)
		} else if strings.HasPrefix(line, "event: done") {
			return
		}
	}
}
