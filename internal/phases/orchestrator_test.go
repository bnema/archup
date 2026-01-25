package phases

import (
	"fmt"
	"testing"

	"github.com/bnema/archup/internal/config"
)

// testPhase is a simple Phase implementation for testing orchestrator
type testPhase struct {
	*BasePhase
	preCheckErr  error
	executeErr   error
	postCheckErr error
	rollbackErr  error
	canSkip      bool
}

func newTestPhase(name string, cfg *config.Config) *testPhase {
	return &testPhase{
		BasePhase: NewBasePhase(name, "Test "+name, cfg, nil),
		canSkip:   false,
	}
}

func (p *testPhase) PreCheck() error {
	return p.preCheckErr
}

func (p *testPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	if p.executeErr != nil {
		return PhaseResult{
			Success: false,
			Error:   p.executeErr,
		}
	}
	return PhaseResult{
		Success: true,
		Message: "completed",
	}
}

func (p *testPhase) PostCheck() error {
	return p.postCheckErr
}

func (p *testPhase) Rollback() error {
	return p.rollbackErr
}

func (p *testPhase) CanSkip() bool {
	return p.canSkip
}

// TestOrchestratorRegistration tests phase registration
func TestOrchestratorRegistration(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	t.Run("register valid phase", func(t *testing.T) {
		phase := newTestPhase("test-phase", cfg)
		err := orch.RegisterPhase(phase)
		if err != nil {
			t.Errorf("Failed to register valid phase: %v", err)
		}

		phases := orch.Phases()
		if len(phases) != 1 {
			t.Errorf("Expected 1 phase, got %d", len(phases))
		}
	})

	t.Run("register nil phase", func(t *testing.T) {
		orch2 := NewOrchestrator(cfg, "test.log")
		err := orch2.RegisterPhase(nil)
		if err == nil {
			t.Error("Expected error for nil phase, got nil")
		}
	})

	t.Run("register duplicate phase name", func(t *testing.T) {
		orch2 := NewOrchestrator(cfg, "test.log")
		phase1 := newTestPhase("duplicate", cfg)
		phase2 := newTestPhase("duplicate", cfg)

		if err := orch2.RegisterPhase(phase1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		err := orch2.RegisterPhase(phase2)
		if err == nil {
			t.Error("Expected error for duplicate phase name, got nil")
		}
	})
}

// TestOrchestratorCurrentPhase tests CurrentPhase navigation
func TestOrchestratorCurrentPhase(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Before execution, current should be nil
	if orch.CurrentPhase() != nil {
		t.Error("CurrentPhase should be nil before execution")
	}

	// After executing, current should be set
	if err := orch.ExecutePhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	current := orch.CurrentPhase()
	if current == nil {
		t.Error("CurrentPhase should not be nil after execution")
	}
	if current.Name() != "test" {
		t.Errorf("CurrentPhase name = %q, want %q", current.Name(), "test")
	}
}

// TestOrchestratorNextPhase tests NextPhase logic
func TestOrchestratorNextPhase(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase1 := newTestPhase("phase-1", cfg)
	phase2 := newTestPhase("phase-2", cfg)
	phase3 := newTestPhase("phase-3", cfg)

	if err := orch.RegisterPhase(phase1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase2); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase3); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Next should be phase-1
	next := orch.NextPhase()
	if next == nil || next.Name() != "phase-1" {
		t.Errorf("NextPhase = %v, want phase-1", next)
	}

	// Mark phase-1 completed
	phase1.SetStatus(StatusCompleted)

	// Next should be phase-2
	next = orch.NextPhase()
	if next == nil || next.Name() != "phase-2" {
		t.Errorf("NextPhase = %v, want phase-2", next)
	}

	// Mark phase-2 skipped
	phase2.SetStatus(StatusSkipped)

	// Next should be phase-3
	next = orch.NextPhase()
	if next == nil || next.Name() != "phase-3" {
		t.Errorf("NextPhase = %v, want phase-3", next)
	}

	// Mark all done
	phase3.SetStatus(StatusCompleted)

	// Next should be nil
	next = orch.NextPhase()
	if next != nil {
		t.Errorf("NextPhase = %v, want nil when all done", next)
	}
}

// TestOrchestratorExecutePhaseSuccess tests successful phase execution
func TestOrchestratorExecutePhaseSuccess(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecutePhase(phase)
	if err != nil {
		t.Errorf("ExecutePhase failed: %v", err)
	}

	if phase.Status() != StatusCompleted {
		t.Errorf("Status = %v, want %v", phase.Status(), StatusCompleted)
	}
}

// TestOrchestratorExecutePhasePreCheckFails tests PreCheck failure
func TestOrchestratorExecutePhasePreCheckFails(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	phase.preCheckErr = fmt.Errorf("precondition failed")
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecutePhase(phase)
	if err == nil {
		t.Error("Expected error from failed PreCheck")
	}

	if phase.Status() != StatusFailed {
		t.Errorf("Status = %v, want %v", phase.Status(), StatusFailed)
	}
}

// TestOrchestratorExecutePhaseExecuteFails tests Execute failure with Rollback
func TestOrchestratorExecutePhaseExecuteFails(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	phase.executeErr = fmt.Errorf("execution failed")
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecutePhase(phase)
	if err == nil {
		t.Error("Expected error from failed Execute")
	}

	if phase.Status() != StatusFailed {
		t.Errorf("Status = %v, want %v", phase.Status(), StatusFailed)
	}
}

// TestOrchestratorExecutePhaseRollbackFails tests Rollback failure
func TestOrchestratorExecutePhaseRollbackFails(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	phase.executeErr = fmt.Errorf("execution failed")
	phase.rollbackErr = fmt.Errorf("rollback also failed")
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecutePhase(phase)
	if err == nil {
		t.Error("Expected error from failed Rollback")
	}

	if phase.Status() != StatusFailed {
		t.Errorf("Status = %v, want %v", phase.Status(), StatusFailed)
	}
}

// TestOrchestratorExecutePhasePostCheckFails tests PostCheck failure
func TestOrchestratorExecutePhasePostCheckFails(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	phase.postCheckErr = fmt.Errorf("validation failed")
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecutePhase(phase)
	if err == nil {
		t.Error("Expected error from failed PostCheck")
	}

	if phase.Status() != StatusFailed {
		t.Errorf("Status = %v, want %v", phase.Status(), StatusFailed)
	}
}

// TestOrchestratorSkipPhase tests skipping phases
func TestOrchestratorSkipPhase(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	t.Run("skip skippable phase", func(t *testing.T) {
		phase := newTestPhase("skippable", cfg)
		phase.canSkip = true
		if err := orch.RegisterPhase(phase); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		err := orch.SkipPhase(phase)
		if err != nil {
			t.Errorf("SkipPhase failed: %v", err)
		}

		if phase.Status() != StatusSkipped {
			t.Errorf("Status = %v, want %v", phase.Status(), StatusSkipped)
		}
	})

	t.Run("skip non-skippable phase", func(t *testing.T) {
		orch2 := NewOrchestrator(cfg, "test.log")
		phase := newTestPhase("non-skippable", cfg)
		phase.canSkip = false
		if err := orch2.RegisterPhase(phase); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		err := orch2.SkipPhase(phase)
		if err == nil {
			t.Error("Expected error when skipping non-skippable phase")
		}
	})
}

// TestOrchestratorProgress tests progress calculation
func TestOrchestratorProgress(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase1 := newTestPhase("phase-1", cfg)
	phase2 := newTestPhase("phase-2", cfg)
	phase3 := newTestPhase("phase-3", cfg)

	if err := orch.RegisterPhase(phase1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase2); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase3); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Initially all pending
	completed, total := orch.Progress()
	if total != 3 {
		t.Errorf("Total = %d, want 3", total)
	}
	if completed != 0 {
		t.Errorf("Completed = %d, want 0", completed)
	}

	// Mark one completed, one skipped
	phase1.SetStatus(StatusCompleted)
	phase2.SetStatus(StatusSkipped)

	completed, _ = orch.Progress()
	if completed != 2 {
		t.Errorf("Completed = %d, want 2 (completed + skipped)", completed)
	}
}

// TestOrchestratorIsComplete tests completion detection
func TestOrchestratorIsComplete(t *testing.T) {
	cfg := config.NewConfig("test")

	t.Run("all completed", func(t *testing.T) {
		orch := NewOrchestrator(cfg, "test.log")
		phase := newTestPhase("test", cfg)
		phase.SetStatus(StatusCompleted)
		if err := orch.RegisterPhase(phase); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !orch.IsComplete() {
			t.Error("IsComplete should be true when all phases completed")
		}
	})

	t.Run("with skipped", func(t *testing.T) {
		orch := NewOrchestrator(cfg, "test.log")
		phase := newTestPhase("test", cfg)
		phase.SetStatus(StatusSkipped)
		if err := orch.RegisterPhase(phase); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !orch.IsComplete() {
			t.Error("IsComplete should be true when all phases skipped")
		}
	})

	t.Run("with pending", func(t *testing.T) {
		orch := NewOrchestrator(cfg, "test.log")
		phase := newTestPhase("test", cfg)
		if err := orch.RegisterPhase(phase); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if orch.IsComplete() {
			t.Error("IsComplete should be false with pending phases")
		}
	})
}

// TestOrchestratorHasFailed tests failure detection
func TestOrchestratorHasFailed(t *testing.T) {
	cfg := config.NewConfig("test")

	t.Run("no failures", func(t *testing.T) {
		orch := NewOrchestrator(cfg, "test.log")
		phase := newTestPhase("test", cfg)
		phase.SetStatus(StatusCompleted)
		if err := orch.RegisterPhase(phase); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if orch.HasFailed() {
			t.Error("HasFailed should be false when no phases failed")
		}
	})

	t.Run("with failure", func(t *testing.T) {
		orch := NewOrchestrator(cfg, "test.log")
		phase := newTestPhase("test", cfg)
		phase.SetStatus(StatusFailed)
		if err := orch.RegisterPhase(phase); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !orch.HasFailed() {
			t.Error("HasFailed should be true when phase failed")
		}
	})
}

// TestOrchestratorReset tests Reset functionality
func TestOrchestratorReset(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase1 := newTestPhase("phase-1", cfg)
	phase2 := newTestPhase("phase-2", cfg)

	phase1.SetStatus(StatusCompleted)
	phase2.SetStatus(StatusFailed)

	if err := orch.RegisterPhase(phase1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase2); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Execute to set currentIdx
	if err := orch.ExecutePhase(phase1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Reset
	orch.Reset()

	// Verify currentIdx reset
	if orch.CurrentPhase() != nil {
		t.Error("CurrentPhase should be nil after Reset")
	}

	// Verify statuses reset
	if phase1.Status() != StatusPending {
		t.Errorf("phase1 Status = %v, want %v", phase1.Status(), StatusPending)
	}
	if phase2.Status() != StatusPending {
		t.Errorf("phase2 Status = %v, want %v", phase2.Status(), StatusPending)
	}
}

// TestOrchestratorExecuteAll tests full sequence execution
func TestOrchestratorExecuteAll(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase1 := newTestPhase("phase-1", cfg)
	phase2 := newTestPhase("phase-2", cfg)
	phase3 := newTestPhase("phase-3", cfg)

	if err := orch.RegisterPhase(phase1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase2); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase3); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecuteAll()
	if err != nil {
		t.Errorf("ExecuteAll failed: %v", err)
	}

	// Verify all completed
	if phase1.Status() != StatusCompleted {
		t.Errorf("phase1 Status = %v, want %v", phase1.Status(), StatusCompleted)
	}
	if phase2.Status() != StatusCompleted {
		t.Errorf("phase2 Status = %v, want %v", phase2.Status(), StatusCompleted)
	}
	if phase3.Status() != StatusCompleted {
		t.Errorf("phase3 Status = %v, want %v", phase3.Status(), StatusCompleted)
	}
}

// TestOrchestratorExecuteAllStopsOnFailure tests that ExecuteAll stops on first failure
func TestOrchestratorExecuteAllStopsOnFailure(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase1 := newTestPhase("phase-1", cfg)
	phase2 := newTestPhase("phase-2", cfg)
	phase3 := newTestPhase("phase-3", cfg)

	// Make phase-2 fail
	phase2.executeErr = fmt.Errorf("execution failed")

	if err := orch.RegisterPhase(phase1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase2); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := orch.RegisterPhase(phase3); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecuteAll()
	if err == nil {
		t.Error("Expected error from ExecuteAll when phase-2 fails")
	}

	// Verify phase-1 completed
	if phase1.Status() != StatusCompleted {
		t.Errorf("phase1 Status = %v, want %v", phase1.Status(), StatusCompleted)
	}

	// Verify phase-2 failed
	if phase2.Status() != StatusFailed {
		t.Errorf("phase2 Status = %v, want %v", phase2.Status(), StatusFailed)
	}

	// Verify phase-3 never executed
	if phase3.Status() != StatusPending {
		t.Errorf("phase3 Status = %v, want %v (should not execute)", phase3.Status(), StatusPending)
	}
}

// TestOrchestratorExecuteNext tests ExecuteNext
func TestOrchestratorExecuteNext(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecuteNext()
	if err != nil {
		t.Errorf("ExecuteNext failed: %v", err)
	}

	if phase.Status() != StatusCompleted {
		t.Errorf("Status = %v, want %v", phase.Status(), StatusCompleted)
	}
}

// TestOrchestratorExecuteNextNoPending tests ExecuteNext with no pending phases
func TestOrchestratorExecuteNextNoPending(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	phase.SetStatus(StatusCompleted)
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecuteNext()
	if err == nil {
		t.Error("Expected error when no pending phases")
	}
}

// TestOrchestratorPhaseDurations tests duration tracking
func TestOrchestratorPhaseDurations(t *testing.T) {
	cfg := config.NewConfig("test")
	orch := NewOrchestrator(cfg, "test.log")

	phase := newTestPhase("test", cfg)
	if err := orch.RegisterPhase(phase); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := orch.ExecutePhase(phase)
	if err != nil {
		t.Fatalf("ExecutePhase failed: %v", err)
	}

	durations := orch.GetPhaseDurations()
	duration, exists := durations["test"]
	if !exists {
		t.Error("Expected duration for 'test' phase")
	}
	if duration == 0 {
		t.Error("Duration should be greater than 0")
	}
}
