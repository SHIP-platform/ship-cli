package api

import (
	"fmt"
	"net/url"
	"strings"
)

const DefaultWebSocketBase = "wss://console.ship-platform.com"

// BuildPortForwardURL constructs the public port-forward WebSocket endpoint.
// The WebSocket origin is intentionally independent from the REST API origin.
func BuildPortForwardURL(baseURL, applicationID, token string) (string, error) {
	if applicationID == "" {
		return "", fmt.Errorf("application ID is required")
	}

	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse WebSocket server URL: %w", err)
	}
	if endpoint.Scheme != "ws" && endpoint.Scheme != "wss" {
		return "", fmt.Errorf("WebSocket server URL must use ws or wss")
	}
	if endpoint.Host == "" {
		return "", fmt.Errorf("WebSocket server URL must include a host")
	}
	if endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return "", fmt.Errorf("WebSocket server URL must not include user info, query, or fragment")
	}

	basePath := strings.TrimRight(endpoint.Path, "/")
	baseRawPath := strings.TrimRight(endpoint.EscapedPath(), "/")
	endpoint.Path = basePath + "/ws/portforward/" + applicationID
	endpoint.RawPath = baseRawPath + "/ws/portforward/" + url.PathEscape(applicationID)
	query := endpoint.Query()
	query.Set("token", token)
	endpoint.RawQuery = query.Encode()

	return endpoint.String(), nil
}
