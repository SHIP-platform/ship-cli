package api

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildPortForwardURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		base    string
		appID   string
		token   string
		wantURL string
	}{
		{
			name:    "production default",
			base:    DefaultWebSocketBase,
			appID:   "app-123",
			token:   "ship_pat_token",
			wantURL: "wss://console.ship-platform.com/ws/portforward/app-123?token=ship_pat_token",
		},
		{
			name:    "custom base and trailing slash",
			base:    "ws://127.0.0.1:3000/edge/",
			appID:   "app-123",
			token:   "token",
			wantURL: "ws://127.0.0.1:3000/edge/ws/portforward/app-123?token=token",
		},
		{
			name:    "encoded path and query values",
			base:    "wss://example.test",
			appID:   "app/id with space",
			token:   "token&scope=admin",
			wantURL: "wss://example.test/ws/portforward/app%2Fid%20with%20space?token=token%26scope%3Dadmin",
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := BuildPortForwardURL(testCase.base, testCase.appID, testCase.token)
			if err != nil {
				t.Fatalf("BuildPortForwardURL: %v", err)
			}
			if got != testCase.wantURL {
				t.Fatalf("URL = %q, want %q", got, testCase.wantURL)
			}
			parsed, err := url.Parse(got)
			if err != nil {
				t.Fatalf("parse result: %v", err)
			}
			if parsed.Query().Get("token") != testCase.token {
				t.Fatalf("decoded token = %q, want %q", parsed.Query().Get("token"), testCase.token)
			}
		})
	}
}

func TestBuildPortForwardURLRejectsInvalidBase(t *testing.T) {
	t.Parallel()

	for _, base := range []string{
		"https://console.ship-platform.com",
		"wss:///missing-host",
		"wss://user@example.test",
		"wss://example.test?source=test",
		"wss://example.test#fragment",
	} {
		base := base
		t.Run(strings.ReplaceAll(base, "/", "_"), func(t *testing.T) {
			t.Parallel()
			if _, err := BuildPortForwardURL(base, "app-123", "token"); err == nil {
				t.Fatalf("BuildPortForwardURL(%q) should fail", base)
			}
		})
	}
}

func TestBuildPortForwardURLRequiresApplicationID(t *testing.T) {
	t.Parallel()
	if _, err := BuildPortForwardURL(DefaultWebSocketBase, "", "token"); err == nil {
		t.Fatal("BuildPortForwardURL should require an application ID")
	}
}
