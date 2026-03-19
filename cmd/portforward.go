package cmd

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	localPort  int
	targetPort int
)

var portForwardCmd = &cobra.Command{
	Use:   "port-forward [APP_ID]",
	Short: "Forward one or more local ports to a pod",
	Long:  `Forward one or more local ports to a pod running in the SHIP Platform.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appID := args[0]

		if token == "" {
			log.Fatal("Error: --token is required")
		}

		listenAddr := fmt.Sprintf("localhost:%d", localPort)
		l, err := net.Listen("tcp", listenAddr)
		if err != nil {
			log.Fatalf("Failed to listen on %s: %v", listenAddr, err)
		}
		defer l.Close()

		fmt.Printf("Forwarding from 127.0.0.1:%d -> %d\n", localPort, targetPort)
		fmt.Printf("Forwarding from [::1]:%d -> %d\n", localPort, targetPort)

		// Handle graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println("\nShutting down port-forward...")
			l.Close()
			os.Exit(0)
		}()

		for {
			conn, err := l.Accept()
			if err != nil {
				// If listener is closed, exit loop
				return
			}
			fmt.Println("Handling connection for", localPort)
			go handleConnection(conn, appID)
		}
	},
}

func init() {
	rootCmd.AddCommand(portForwardCmd)

	portForwardCmd.Flags().IntVarP(&localPort, "local-port", "l", 8080, "Local port to listen on")
	portForwardCmd.Flags().IntVarP(&targetPort, "target-port", "t", 80, "Target port on the pod")
	portForwardCmd.MarkFlagRequired("token")
}

func handleConnection(localConn net.Conn, appID string) {
	defer localConn.Close()

	wsURL := fmt.Sprintf("%s/ws/portforward/%s?port=%d&token=%s", apiServer, appID, targetPort, token)

	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("Failed to connect to backend: %v", err)
		return
	}
	defer ws.Close()

	errCh := make(chan error, 2)

	// Local TCP -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := localConn.Read(buf)
			if n > 0 {
				if wErr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); wErr != nil {
					errCh <- wErr
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					errCh <- err
				} else {
					errCh <- nil
				}
				return
			}
		}
	}()

	// WebSocket -> Local TCP
	go func() {
		for {
			mt, data, err := ws.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if mt == websocket.BinaryMessage {
				if _, wErr := localConn.Write(data); wErr != nil {
					errCh <- wErr
					return
				}
			}
		}
	}()

	<-errCh
	fmt.Println("Connection closed")
}
