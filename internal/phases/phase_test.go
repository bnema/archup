package phases

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
)

// TestPhaseStatusString tests the String() method of PhaseStatus enum
func TestPhaseStatusString(t *testing.T) {
	tests := []struct {
		status   PhaseStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusRunning, "running"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{StatusSkipped, "skipped"},
		{PhaseStatus(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.expected {
				t.Errorf("PhaseStatus.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestNewBasePhase verifies BasePhase initialization
func TestNewBasePhase(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")

	phase := NewBasePhase("test-phase", "Test phase description", cfg, log)

	if phase.Name() != "test-phase" {
		t.Errorf("Name() = %q, want %q", phase.Name(), "test-phase")
	}

	if phase.Description() != "Test phase description" {
		t.Errorf("Description() = %q, want %q", phase.Description(), "Test phase description")
	}

	if phase.Status() != StatusPending {
		t.Errorf("Status() = %v, want %v", phase.Status(), StatusPending)
	}

	if phase.Config() != cfg {
		t.Errorf("Config() returned different instance")
	}

	if phase.Logger() != log {
		t.Errorf("Logger() returned different instance")
	}
}

// TestBasePhaseSetStatus verifies status updates
func TestBasePhaseSetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("test", "test", cfg, log)

	statuses := []PhaseStatus{
		StatusPending,
		StatusRunning,
		StatusCompleted,
		StatusFailed,
		StatusSkipped,
	}

	for _, status := range statuses {
		t.Run(status.String(), func(t *testing.T) {
			phase.SetStatus(status)
			if phase.Status() != status {
				t.Errorf("Status() = %v, want %v", phase.Status(), status)
			}
		})
	}
}

// TestBasePhaseDefaultPreCheck verifies default PreCheck implementation
func TestBasePhaseDefaultPreCheck(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	tests := []struct {
		name      string
		cfg       *config.Config
		wantError bool
	}{
		{
			name:      "with valid config",
			cfg:       config.NewConfig("test"),
			wantError: false,
		},
		{
			name:      "with nil config",
			cfg:       nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase := NewBasePhase("test", "test", tt.cfg, log)
			err := phase.PreCheck()

			if (err != nil) != tt.wantError {
				t.Errorf("PreCheck() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestBasePhaseDefaultPostCheck verifies default PostCheck (always succeeds)
func TestBasePhaseDefaultPostCheck(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("test", "test", cfg, log)

	err := phase.PostCheck()
	if err != nil {
		t.Errorf("PostCheck() error = %v, want nil", err)
	}
}

// TestBasePhaseDefaultRollback verifies default Rollback returns error
func TestBasePhaseDefaultRollback(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("my-phase", "test", cfg, log)

	err := phase.Rollback()
	if err == nil {
		t.Error("Rollback() error = nil, want error")
	}

	expectedMsg := "rollback not implemented for phase: my-phase"
	if err.Error() != expectedMsg {
		t.Errorf("Rollback() error = %q, want %q", err.Error(), expectedMsg)
	}
}

// TestBasePhaseDefaultCanSkip verifies default CanSkip returns false
func TestBasePhaseDefaultCanSkip(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("test", "test", cfg, log)

	if phase.CanSkip() {
		t.Error("CanSkip() = true, want false")
	}
}

// TestSendProgress verifies progress updates sent to channel
func TestSendProgress(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("test-phase", "test", cfg, log)

	progressChan := make(chan ProgressUpdate, 5)

	phase.SendProgress(progressChan, "step-1", 1, 3)
	phase.SendProgress(progressChan, "step-2", 2, 3)

	// Verify first update
	update1 := <-progressChan
	if update1.PhaseName != "test-phase" {
		t.Errorf("PhaseName = %q, want %q", update1.PhaseName, "test-phase")
	}
	if update1.Step != "step-1" {
		t.Errorf("Step = %q, want %q", update1.Step, "step-1")
	}
	if update1.Current != 1 || update1.Total != 3 {
		t.Errorf("Progress = %d/%d, want 1/3", update1.Current, update1.Total)
	}
	if update1.IsComplete || update1.IsError {
		t.Error("Flags should be false for regular progress update")
	}

	// Verify second update
	update2 := <-progressChan
	if update2.Step != "step-2" {
		t.Errorf("Step = %q, want %q", update2.Step, "step-2")
	}
	if update2.Current != 2 {
		t.Errorf("Current = %d, want 2", update2.Current)
	}
}

// TestSendProgressWithNilChannel verifies no panic with nil channel
func TestSendProgressWithNilChannel(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("test", "test", cfg, log)

	// Should not panic with nil channel
	phase.SendProgress(nil, "step", 1, 2)
}

// TestSendComplete verifies completion signal
func TestSendComplete(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("test-phase", "test", cfg, log)

	progressChan := make(chan ProgressUpdate, 1)

	phase.SendComplete(progressChan, "Installation complete")

	update := <-progressChan
	if !update.IsComplete {
		t.Error("IsComplete = false, want true")
	}
	if update.IsError {
		t.Error("IsError = true, want false")
	}
	if update.Step != "Installation complete" {
		t.Errorf("Step = %q, want %q", update.Step, "Installation complete")
	}
}

// TestSendError verifies error signal
func TestSendError(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("test-phase", "test", cfg, log)

	progressChan := make(chan ProgressUpdate, 1)

	testErr := "test error message"
	err := fmt.Errorf("%s", testErr)
	phase.SendError(progressChan, err)

	update := <-progressChan
	if !update.IsComplete {
		t.Error("IsComplete = false, want true")
	}
	if !update.IsError {
		t.Error("IsError = false, want true")
	}
	if update.ErrorMsg != testErr {
		t.Errorf("ErrorMsg = %q, want %q", update.ErrorMsg, testErr)
	}
}

// TestSendOutput verifies output line sending
func TestSendOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	log, _ := logger.New(logPath, false)
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBasePhase("test-phase", "test", cfg, log)

	progressChan := make(chan ProgressUpdate, 1)

	outputLine := "output from command"
	phase.SendOutput(progressChan, outputLine)

	update := <-progressChan
	if update.OutputLine != outputLine {
		t.Errorf("OutputLine = %q, want %q", update.OutputLine, outputLine)
	}
	if update.IsComplete || update.IsError {
		t.Error("Completion/Error flags should be false for output update")
	}
}
