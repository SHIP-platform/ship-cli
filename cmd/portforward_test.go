package cmd

import (
	"testing"

	"ship-cli/api"
)

func TestPortForwardCommandUsesLocalPortOnly(t *testing.T) {
	if portForwardCmd.Flags().Lookup("local-port") == nil {
		t.Fatal("port-forward must expose --local-port")
	}
	if portForwardCmd.Flags().Lookup("target-port") != nil {
		t.Fatal("port-forward must not expose --target-port")
	}

	flag := rootCmd.PersistentFlags().Lookup("websocket-server")
	if flag == nil {
		t.Fatal("root command must expose --websocket-server")
	}
	if flag.DefValue != api.DefaultWebSocketBase {
		t.Fatalf("--websocket-server default = %q, want %q", flag.DefValue, api.DefaultWebSocketBase)
	}
}
