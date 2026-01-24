package views

import (
	"strings"
	"testing"

	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/interfaces/tui/models"
)

func TestRenderProgress(t *testing.T) {
	pm := models.NewProgressModel()

	// Update with some progress
	update := &dto.ProgressUpdate{
		Phase:           "Partitioning",
		PhaseNumber:     2,
		TotalPhases:     7,
		ProgressPercent: 25,
		Message:         "Partitioning disk...",
		IsError:         false,
	}
	pm.UpdateProgress(update)

	// Render the progress
	output := RenderProgress(pm)

	// Verify key content is present
	checks := []string{
		"Installation Progress", // Title
		"Phase 2/7",             // Phase info
		"Partitioning",          // Phase name
		"Partitioning disk...",  // Message
		"[",                     // Progress bar start
		"]",                     // Progress bar end
		"25%",                   // Progress percentage
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Expected progress output to contain '%s', but it didn't", check)
		}
	}

	// Verify output is not empty
	if len(output) == 0 {
		t.Error("Progress output should not be empty")
	}

	t.Logf("Progress rendered successfully with %d characters", len(output))
}

func TestRenderProgressError(t *testing.T) {
	pm := models.NewProgressModel()

	// Update with error message
	update := &dto.ProgressUpdate{
		Phase:           "Base Installation",
		PhaseNumber:     3,
		TotalPhases:     7,
		ProgressPercent: 40,
		Message:         "Failed to install base packages",
		IsError:         true,
	}
	pm.UpdateProgress(update)

	// Render the progress
	output := RenderProgress(pm)

	// Verify error message is present
	if !strings.Contains(output, "Failed to install base packages") {
		t.Error("Expected error message in output")
	}

	if !strings.Contains(output, "Base Installation") {
		t.Error("Expected phase name in output")
	}
}

func TestProgressMessageHistory(t *testing.T) {
	pm := models.NewProgressModel()

	// Add several messages
	messages := []string{
		"Starting installation...",
		"Checking system...",
		"Partitioning disk...",
		"Installing base system...",
	}

	for i, msg := range messages {
		update := &dto.ProgressUpdate{
			Phase:           "Test Phase",
			PhaseNumber:     1,
			TotalPhases:     7,
			ProgressPercent: (i + 1) * 20,
			Message:         msg,
			IsError:         false,
		}
		pm.UpdateProgress(update)
	}

	// Render the progress
	output := RenderProgress(pm)

	// Verify history section exists
	if !strings.Contains(output, "Recent activities") {
		t.Error("Expected recent activities section in output")
	}

	// Verify recent messages are shown
	if !strings.Contains(output, "Partitioning disk...") {
		t.Error("Expected recent message in output")
	}
}

func TestProgressBar(t *testing.T) {
	pm := models.NewProgressModel()

	testCases := []struct {
		percent int
		name    string
	}{
		{0, "empty"},
		{25, "quarter"},
		{50, "half"},
		{75, "three-quarters"},
		{100, "full"},
	}

	for _, tc := range testCases {
		update := &dto.ProgressUpdate{
			Phase:           "Test",
			PhaseNumber:     1,
			TotalPhases:     7,
			ProgressPercent: tc.percent,
			Message:         "Testing",
			IsError:         false,
		}
		pm.UpdateProgress(update)

		output := RenderProgress(pm)

		if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
			t.Errorf("Progress bar for %s should contain [ and ]", tc.name)
		}

		if !strings.Contains(output, "100%") && tc.percent == 100 {
			t.Errorf("Progress bar for %s should show 100%%", tc.name)
		}
	}
}

func TestProgressModelGetters(t *testing.T) {
	pm := models.NewProgressModel()

	// Initial state
	if phase, total := pm.GetPhaseProgress(); phase != 0 || total != 0 {
		t.Error("Initial phase progress should be 0/0")
	}

	if pm.GetProgressPercent() != 0 {
		t.Error("Initial progress percent should be 0")
	}

	if pm.GetCurrentPhase() != "" {
		t.Error("Initial current phase should be empty")
	}

	// Update
	update := &dto.ProgressUpdate{
		Phase:           "Partitioning",
		PhaseNumber:     2,
		TotalPhases:     7,
		ProgressPercent: 30,
		Message:         "Working...",
		IsError:         false,
	}
	pm.UpdateProgress(update)

	// Verify getters work
	if phase, total := pm.GetPhaseProgress(); phase != 2 || total != 7 {
		t.Errorf("Expected 2/7 phases, got %d/%d", phase, total)
	}

	if pm.GetProgressPercent() != 30 {
		t.Errorf("Expected 30%%, got %d%%", pm.GetProgressPercent())
	}

	if pm.GetCurrentPhase() != "Partitioning" {
		t.Error("Expected current phase to be 'Partitioning'")
	}

	if pm.GetMessage() != "Working..." {
		t.Error("Expected message to be 'Working...'")
	}

	if pm.IsError() {
		t.Error("Expected IsError to be false")
	}
}

func BenchmarkRenderProgress(b *testing.B) {
	pm := models.NewProgressModel()

	update := &dto.ProgressUpdate{
		Phase:           "Test",
		PhaseNumber:     1,
		TotalPhases:     7,
		ProgressPercent: 50,
		Message:         "Testing",
		IsError:         false,
	}
	pm.UpdateProgress(update)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderProgress(pm)
	}
}
