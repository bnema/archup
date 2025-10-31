package installation

import (
	"context"
	"testing"
)

func TestNewInstallation_Valid(t *testing.T) {
	inst, err := NewInstallation(
		"myhost",
		"user",
		"/dev/sda",
		"none",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if inst.ID() == "" {
		t.Error("expected non-empty ID")
	}

	if inst.Hostname() != "myhost" {
		t.Errorf("expected hostname 'myhost', got '%s'", inst.Hostname())
	}

	if inst.Username() != "user" {
		t.Errorf("expected username 'user', got '%s'", inst.Username())
	}

	if inst.TargetDisk() != "/dev/sda" {
		t.Errorf("expected target disk '/dev/sda', got '%s'", inst.TargetDisk())
	}

	if inst.EncryptionType() != "none" {
		t.Errorf("expected encryption 'none', got '%s'", inst.EncryptionType())
	}

	if inst.State() != StateNotStarted {
		t.Errorf("expected initial state StateNotStarted, got %v", inst.State())
	}

	if inst.IsStarted() {
		t.Error("expected IsStarted to be false")
	}

	if inst.IsCompleted() {
		t.Error("expected IsCompleted to be false")
	}
}

func TestNewInstallation_InvalidHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		wantErr  bool
	}{
		{"empty hostname", "", true},
		{"too long hostname", "a" + string(make([]byte, 100)), true},
		{"valid hostname", "myhost", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewInstallation(tt.hostname, "user", "/dev/sda", "none")
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, want error=%v", err, tt.wantErr)
			}
		})
	}
}

func TestNewInstallation_InvalidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"empty username", "", true},
		{"too long username", "a" + string(make([]byte, 100)), true},
		{"valid username", "user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewInstallation("host", tt.username, "/dev/sda", "none")
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, want error=%v", err, tt.wantErr)
			}
		})
	}
}

func TestNewInstallation_InvalidEncryption(t *testing.T) {
	tests := []struct {
		name       string
		encryption string
		wantErr    bool
	}{
		{"invalid encryption", "invalid", true},
		{"none", "none", false},
		{"luks", "luks", false},
		{"luks-lvm", "luks-lvm", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewInstallation("host", "user", "/dev/sda", tt.encryption)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, want error=%v", err, tt.wantErr)
			}
		})
	}
}

func TestStart_Success(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()

	if err := inst.Start(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if inst.State() != StatePreflightChecks {
		t.Errorf("expected state StatePreflightChecks, got %v", inst.State())
	}

	if inst.StartedAt() == nil {
		t.Error("expected StartedAt to be set")
	}

	if inst.IsStarted() != true {
		t.Error("expected IsStarted to be true")
	}

	// Check event was recorded
	events := inst.UncommittedEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	if events[0].EventType() != "InstallationStarted" {
		t.Errorf("expected InstallationStarted event, got %s", events[0].EventType())
	}
}

func TestStart_AlreadyStarted(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()

	inst.Start(ctx)

	// Try to start again
	err := inst.Start(ctx)
	if err != ErrInstallationAlreadyStarted {
		t.Errorf("expected ErrInstallationAlreadyStarted, got %v", err)
	}
}

func TestStateTransition_ValidSequence(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()

	// Start installation
	inst.Start(ctx)
	inst.events = []DomainEvent{} // Clear events

	// Test valid sequence
	phases := []State{
		StatePreflightChecks,
		StateDiskPartitioning,
		StateBaseInstallation,
		StateSystemConfiguration,
		StateBootloaderSetup,
		StateRepositorySetup,
		StatePostInstallation,
		StateCompleted,
	}

	for i := 1; i < len(phases); i++ {
		if err := inst.TransitionToNextPhase(); err != nil {
			t.Errorf("phase %d: unexpected error: %v", i, err)
		}

		if inst.State() != phases[i] {
			t.Errorf("expected state %v, got %v", phases[i], inst.State())
		}
	}

	if !inst.IsCompleted() {
		t.Error("expected installation to be completed")
	}

	if !inst.IsSuccessful() {
		t.Error("expected installation to be successful")
	}
}

func TestCompleteCurrentPhase(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()

	inst.Start(ctx)
	inst.events = []DomainEvent{} // Clear events

	// Complete first phase
	if err := inst.CompleteCurrentPhase(30); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if inst.State() != StateDiskPartitioning {
		t.Errorf("expected state StateDiskPartitioning, got %v", inst.State())
	}

	// Check events
	events := inst.UncommittedEvents()
	if len(events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(events))
	}

	// First event should be PhaseCompleted
	if events[0].EventType() != "PhaseCompleted" {
		t.Errorf("expected PhaseCompleted event, got %s", events[0].EventType())
	}

	// Second event should be PhaseStarted
	if events[1].EventType() != "PhaseStarted" {
		t.Errorf("expected PhaseStarted event, got %s", events[1].EventType())
	}
}

func TestFailCurrentPhase(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()

	inst.Start(ctx)
	inst.events = []DomainEvent{} // Clear events

	// Fail the phase
	if err := inst.FailCurrentPhase("disk not found", false); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if inst.State() != StateFailed {
		t.Errorf("expected state StateFailed, got %v", inst.State())
	}

	if !inst.IsFailed() {
		t.Error("expected IsFailed to be true")
	}

	if inst.IsSuccessful() {
		t.Error("expected IsSuccessful to be false")
	}

	if inst.IsCompleted() != true {
		t.Error("expected IsCompleted to be true")
	}

	// Check events
	events := inst.UncommittedEvents()
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	if events[0].EventType() != "PhaseFailed" {
		t.Errorf("expected PhaseFailed event, got %s", events[0].EventType())
	}

	if events[1].EventType() != "InstallationFailed" {
		t.Errorf("expected InstallationFailed event, got %s", events[1].EventType())
	}
}

func TestComplete_Success(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()

	inst.Start(ctx)

	// Advance to PostInstallation phase
	for inst.State() != StatePostInstallation {
		inst.TransitionToNextPhase()
	}

	inst.events = []DomainEvent{} // Clear events

	// Complete installation
	if err := inst.Complete(300); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if inst.State() != StateCompleted {
		t.Errorf("expected state StateCompleted, got %v", inst.State())
	}

	if inst.CompletedAt() == nil {
		t.Error("expected CompletedAt to be set")
	}

	// Check event
	events := inst.UncommittedEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	if events[0].EventType() != "InstallationCompleted" {
		t.Errorf("expected InstallationCompleted event, got %s", events[0].EventType())
	}
}

func TestComplete_InvalidState(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()

	inst.Start(ctx)

	// Try to complete from wrong state
	err := inst.Complete(300)
	if err != ErrInvalidPhaseTransition {
		t.Errorf("expected ErrInvalidPhaseTransition, got %v", err)
	}
}

func TestProgressPercentage(t *testing.T) {
	tests := []struct {
		state      State
		percentage int
	}{
		{StateNotStarted, 0},
		{StatePreflightChecks, 10},
		{StateDiskPartitioning, 25},
		{StateBaseInstallation, 40},
		{StateSystemConfiguration, 60},
		{StateBootloaderSetup, 75},
		{StateRepositorySetup, 85},
		{StatePostInstallation, 95},
		{StateCompleted, 100},
		{StateFailed, 0},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if percentage := tt.state.ProgressPercentage(); percentage != tt.percentage {
				t.Errorf("expected %d%%, got %d%%", tt.percentage, percentage)
			}
		})
	}
}

func TestClearEvents(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()

	inst.Start(ctx)

	if len(inst.UncommittedEvents()) == 0 {
		t.Fatal("expected events after start")
	}

	inst.ClearEvents()

	if len(inst.UncommittedEvents()) != 0 {
		t.Error("expected no events after clear")
	}
}

func TestNewInstallation_EncryptionTypes(t *testing.T) {
	validTypes := []string{"none", "luks", "luks-lvm"}

	for _, encType := range validTypes {
		t.Run(encType, func(t *testing.T) {
			inst, err := NewInstallation("host", "user", "/dev/sda", encType)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if inst.EncryptionType() != encType {
				t.Errorf("expected %s, got %s", encType, inst.EncryptionType())
			}
		})
	}
}

func TestInstallation_TimestampAccuracy(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")

	createdAt := inst.CreatedAt()
	if createdAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	// StartedAt should be nil before starting
	if inst.StartedAt() != nil {
		t.Error("expected nil StartedAt before starting")
	}

	// CompletedAt should be nil before completing
	if inst.CompletedAt() != nil {
		t.Error("expected nil CompletedAt before completing")
	}

	ctx := context.Background()
	inst.Start(ctx)

	if inst.StartedAt() == nil {
		t.Error("expected non-nil StartedAt after starting")
	}

	// Verify StartedAt is after CreatedAt
	if inst.StartedAt().Before(createdAt) {
		t.Error("expected StartedAt to be after or equal to CreatedAt")
	}
}

func TestFailCurrentPhase_NotStarted(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")

	err := inst.FailCurrentPhase("error", false)
	if err != ErrInstallationNotStarted {
		t.Errorf("expected ErrInstallationNotStarted, got %v", err)
	}
}

func TestCompleteCurrentPhase_NotStarted(t *testing.T) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")

	err := inst.CompleteCurrentPhase(30)
	if err != ErrInstallationNotStarted {
		t.Errorf("expected ErrInstallationNotStarted, got %v", err)
	}
}

func BenchmarkNewInstallation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewInstallation("host", "user", "/dev/sda", "none")
	}
}

func BenchmarkStateTransition(b *testing.B) {
	inst, _ := NewInstallation("host", "user", "/dev/sda", "none")
	ctx := context.Background()
	inst.Start(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if inst.State() == StatePostInstallation {
			// Reset by creating new instance for a fair benchmark
			inst, _ = NewInstallation("host", "user", "/dev/sda", "none")
			inst.Start(ctx)
		}
		inst.TransitionToNextPhase()
	}
}
