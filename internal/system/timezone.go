package system

import (
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	timezoneAPIURL     = "https://ipapi.co/timezone/"
	timezoneAPITimeout = 5 * time.Second
)

// DetectTimezoneSimple fetches timezone from ipapi.co API without logging.
// Returns detected timezone string or empty string on failure.
func DetectTimezoneSimple() string {
	client := &http.Client{
		Timeout: timezoneAPITimeout,
	}

	resp, err := client.Get(timezoneAPIURL)
	if err != nil {
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		if err := resp.Body.Close(); err != nil {
			return ""
		}
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if err := resp.Body.Close(); err != nil {
			return ""
		}
		return ""
	}

	if err := resp.Body.Close(); err != nil {
		return ""
	}

	timezone := strings.TrimSpace(string(body))
	if !strings.Contains(timezone, "/") {
		return ""
	}

	return timezone
}
