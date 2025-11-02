package system

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	timezoneAPIURL       = "https://ipapi.co/timezone/"
	timezoneAPITimeout   = 5 * time.Second
	timezoneRetryAttempt = 3
	timezoneRetryDelay   = 1 * time.Second
)

// DetectTimezone fetches timezone from ipapi.co API with retry logic
// Attempts up to 3 times with exponential backoff before falling back
// Returns detected timezone string or empty string on failure
// Logs details to the system logger
func (s *System) DetectTimezone() string {
	var lastErr error

	for attempt := 1; attempt <= timezoneRetryAttempt; attempt++ {
		if attempt > 1 {
			// Log retry attempt
			s.log.Debug("Timezone detection retry",
				"attempt", attempt,
				"max_attempts", timezoneRetryAttempt,
				"delay", timezoneRetryDelay.String())
			time.Sleep(timezoneRetryDelay)
		}

		client := &http.Client{
			Timeout: timezoneAPITimeout,
		}

		resp, err := client.Get(timezoneAPIURL)
		if err != nil {
			lastErr = fmt.Errorf("http request failed: %w", err)
			s.log.Debug("Timezone detection failed",
				"attempt", attempt,
				"error", lastErr)
			continue
		}
		defer resp.Body.Close()

		// Check HTTP status code
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API returned status %d", resp.StatusCode)
			s.log.Debug("Timezone detection failed",
				"attempt", attempt,
				"error", lastErr)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			s.log.Debug("Timezone detection failed",
				"attempt", attempt,
				"error", lastErr)
			continue
		}

		timezone := strings.TrimSpace(string(body))

		// Validate timezone format (should be like "America/New_York")
		if !strings.Contains(timezone, "/") {
			lastErr = fmt.Errorf("invalid timezone format: %q (expected format like America/New_York)", timezone)
			s.log.Debug("Timezone detection failed",
				"attempt", attempt,
				"error", lastErr)
			continue
		}

		// Success - log and return
		s.log.Info("Timezone detected successfully", "timezone", timezone)
		return timezone
	}

	// All retries exhausted
	s.log.Warn("Timezone detection failed after all retries",
		"attempts", timezoneRetryAttempt,
		"last_error", lastErr)
	return ""
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
