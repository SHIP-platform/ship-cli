package ui

import (
	"context"
	"strings"
	"testing"

	"ship-cli/api"
)

func TestStartPortForwardRejectsInvalidWebSocketBaseBeforeListening(t *testing.T) {
	client := api.NewClient("https://api.ship-platform.com", "token")
	client.WebSocketBase = "https://console.ship-platform.com"
	ctx, cancel := context.WithCancel(context.Background())

	msg := startPortForward(ctx, cancel, client, api.Application{ID: "app-123", Name: "demo"}, 0)()
	errMsg, ok := msg.(portForwardErrMsg)
	if !ok {
		t.Fatalf("message = %T, want portForwardErrMsg", msg)
	}
	if !strings.Contains(errMsg.err.Error(), "must use ws or wss") {
		t.Fatalf("error = %q, want scheme validation", errMsg.err)
	}
}

func TestPortForwardLabelsDescribeApplicationService(t *testing.T) {
	items := actionItems("app-123", map[string]*PortForwardSession{})
	if len(items) == 0 || !strings.Contains(items[0].(item).desc, "application service") {
		t.Fatalf("start action = %#v, want application service label", items)
	}
}
