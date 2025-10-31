package system

import (
	"fmt"
	"net/http"
	"time"
)

const (
	NetworkCheckURL     = "https://raw.githubusercontent.com"
	NetworkCheckTimeout = 5 * time.Second
)

// CheckNetworkConnectivity verifies internet connectivity
// Returns nil if network is available, error otherwise
func CheckNetworkConnectivity() error {
	client := &http.Client{Timeout: NetworkCheckTimeout}

	resp, err := client.Get(NetworkCheckURL)
	if err != nil {
		return fmt.Errorf("network connectivity check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("network check returned HTTP %d", resp.StatusCode)
	}

	return nil
}
