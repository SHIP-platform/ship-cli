package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [APP_ID]",
	Short: "Stream logs from a deployed application",
	Long:  `Stream live logs from a pod running in the SHIP Platform.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appID := args[0]

		if token == "" {
			log.Fatal("Error: --token is required")
		}

		// Handle graceful shutdown
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
}

func streamLogs(appID string) {
	url := fmt.Sprintf("%s/api/applications/%s/logs/stream", apiServer, appID)
	
	req, err := http.NewRequest("GET", url, nil)
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
		fmt.Print(line)
	}
}
