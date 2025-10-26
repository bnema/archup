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

// DetectTimezone fetches timezone from ipapi.co API
// Returns detected timezone string or empty string on failure
func DetectTimezone() string {
	client := &http.Client{
		Timeout: timezoneAPITimeout,
	}

	resp, err := client.Get(timezoneAPIURL)
	switch {
	case err != nil:
		return ""
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		switch {
		case err != nil:
			return ""
		}

		timezone := strings.TrimSpace(string(body))

		// Validate timezone format (should be like "America/New_York")
		switch {
		case strings.Contains(timezone, "/"):
			return timezone
		default:
			return ""
		}
	default:
		return ""
	}
}

// ValidateTimezone checks if a timezone string is valid
// This is a basic check - full validation would require checking against zoneinfo
func ValidateTimezone(tz string) bool {
	switch {
	case tz == "":
		return false
	case !strings.Contains(tz, "/"):
		return false
	default:
		return true
	}
}
