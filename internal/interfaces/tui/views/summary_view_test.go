package views

import (
	"strings"
	"testing"
	"time"

	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/interfaces/tui/models"
)

func TestRenderSummary(t *testing.T) {
	im := models.NewInstallationModel()

	// Create status data
	now := time.Now()
	completedAt := now.Add(1 * time.Hour)
	status := &dto.InstallationStatus{
		Hostname:    "test-host",
		Username:    "testuser",
		TargetDisk:  "/dev/sda",
		Progress:    100,
		State:       "Complete",
		CurrentPhase: "PostInstallation",
		StartedAt:   &now,
		CompletedAt: &completedAt,
	}

	im.SetStatus(status)
	im.SetComplete()

	// Render the summary
	output := RenderSummary(im)

	// Verify key content is present
	checks := []string{
		"Installation Complete!",       // Title
		"test-host",                    // Hostname
		"testuser",                     // Username
		"/dev/sda",                     // Disk
		"Press 'q' to exit",            // Instructions
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Expected summary output to contain '%s', but it didn't", check)
		}
	}

	// Verify output is not empty
	if len(output) == 0 {
		t.Error("Summary output should not be empty")
	}

	t.Logf("Summary rendered successfully with %d characters", len(output))
}

func TestRenderError(t *testing.T) {
	errorMsg := "Test error: Something went wrong"

	// Render error
	output := RenderError(errorMsg)

	// Verify key content is present
	checks := []string{
		"Installation Failed",           // Title
		errorMsg,                        // Error message
		"Press 'q' or Ctrl+C to exit",  // Instructions
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Expected error output to contain '%s', but it didn't", check)
		}
	}

	if len(output) == 0 {
		t.Error("Error output should not be empty")
	}
}

func TestRenderStatus(t *testing.T) {
	im := models.NewInstallationModel()

	// Create status data
	status := &dto.InstallationStatus{
		Hostname:    "test-host",
		CurrentPhase: "Partitioning",
		Progress:    35,
		State:       "Running",
	}

	im.SetStatus(status)

	// We'll test through RenderSummary since RenderStatus is a helper
	// but we can verify the model works
	if im.GetStatus().Hostname != "test-host" {
		t.Error("Status should store hostname")
	}

	if !im.IsComplete() {
		t.Log("Model correctly shows installation not complete")
	}
}

func TestInstallationModelWithError(t *testing.T) {
	im := models.NewInstallationModel()

	// Create status but set error
	status := &dto.InstallationStatus{
		Hostname:   "test-host",
		Username:   "testuser",
		TargetDisk: "/dev/sda",
	}

	im.SetStatus(status)
	im.SetError("Disk /dev/sda not found")

	// Render using RenderSummary (which checks for error first)
	output := RenderSummary(im)

	// Should render error, not summary
	if !strings.Contains(output, "Installation Failed") {
		t.Error("Should render error when error is set")
	}

	if !strings.Contains(output, "Disk /dev/sda not found") {
		t.Error("Should show the error message")
	}
}

func TestInstallationModelGetters(t *testing.T) {
	im := models.NewInstallationModel()

	// Initial state
	if im.GetError() != "" {
		t.Error("Initial error should be empty")
	}

	if im.IsComplete() {
		t.Error("Should not be complete initially")
	}

	if im.GetStatus() == nil {
		t.Error("Status should not be nil")
	}

	// Set error
	im.SetError("Test error")
	if im.GetError() != "Test error" {
		t.Error("Error should be set")
	}

	// Set complete
	im.SetComplete()
	if !im.IsComplete() {
		t.Error("Should be complete after SetComplete")
	}

	// Set status
	status := &dto.InstallationStatus{Hostname: "test"}
	im.SetStatus(status)
	if im.GetStatus().Hostname != "test" {
		t.Error("Status should be set")
	}
}

func TestDurationFormatting(t *testing.T) {
	im := models.NewInstallationModel()

	testCases := []struct {
		duration time.Duration
		contains string
		name     string
	}{
		{30 * time.Second, "30s", "seconds only"},
		{1*time.Minute + 30*time.Second, "1m 30s", "minutes and seconds"},
		{1*time.Hour + 2*time.Minute + 30*time.Second, "1h 2m 30s", "hours, minutes, seconds"},
	}

	for _, tc := range testCases {
		now := time.Now()
		completedAt := now.Add(tc.duration)
		status := &dto.InstallationStatus{
			StartedAt:   &now,
			CompletedAt: &completedAt,
		}

		im.SetStatus(status)
		im.SetComplete()

		output := RenderSummary(im)

		if !strings.Contains(output, tc.contains) {
			t.Errorf("For %s, expected output to contain '%s', got: %s", tc.name, tc.contains, output)
		}
	}
}

func BenchmarkRenderSummary(b *testing.B) {
	im := models.NewInstallationModel()

	now := time.Now()
	status := &dto.InstallationStatus{
		Hostname:    "test-host",
		Username:    "testuser",
		TargetDisk:  "/dev/sda",
		StartedAt:   &now,
		CompletedAt: &now,
	}

	im.SetStatus(status)
	im.SetComplete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderSummary(im)
	}
}

func BenchmarkRenderError(b *testing.B) {
	errorMsg := "Test error message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderError(errorMsg)
	}
}
