package services

import (
	"testing"
	"time"

	"github.com/bnema/archup/internal/application/dto"
)

func TestProgressTracker_Subscribe(t *testing.T) {
	tracker := NewProgressTracker()
	defer tracker.Close()

	ch := tracker.Subscribe()
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}
}

func TestProgressTracker_Emit(t *testing.T) {
	tracker := NewProgressTracker()
	defer tracker.Close()

	ch := tracker.Subscribe()

	update := &dto.ProgressUpdate{
		Phase:           "Test Phase",
		PhaseNumber:     1,
		TotalPhases:     8,
		ProgressPercent: 10,
		Message:         "Test message",
		IsError:         false,
	}

	go tracker.Emit(update)

	select {
	case received := <-ch:
		if received.Phase != "Test Phase" {
			t.Errorf("expected phase 'Test Phase', got %s", received.Phase)
		}
		if received.Message != "Test message" {
			t.Errorf("expected message 'Test message', got %s", received.Message)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for progress update")
	}
}

func TestProgressTracker_MultipleSubscribers(t *testing.T) {
	tracker := NewProgressTracker()
	defer tracker.Close()

	ch1 := tracker.Subscribe()
	ch2 := tracker.Subscribe()

	update := &dto.ProgressUpdate{
		Phase:   "Test",
		Message: "Test message",
	}

	go tracker.Emit(update)

	// Both channels should receive the update
	select {
	case <-ch1:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout on ch1")
	}

	select {
	case <-ch2:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout on ch2")
	}
}

func TestProgressTracker_EmitPhaseStarted(t *testing.T) {
	tracker := NewProgressTracker()
	defer tracker.Close()

	ch := tracker.Subscribe()

	go tracker.EmitPhaseStarted("Preflight", 1, 8)

	select {
	case update := <-ch:
		if update.Phase != "Preflight" {
			t.Errorf("expected phase 'Preflight', got %s", update.Phase)
		}
		if update.Message != "Starting Preflight" {
			t.Errorf("expected message 'Starting Preflight', got %s", update.Message)
		}
		if update.ProgressPercent != 0 {
			t.Errorf("expected progress 0%%, got %d%%", update.ProgressPercent)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for phase started event")
	}
}

func TestProgressTracker_EmitPhaseCompleted(t *testing.T) {
	tracker := NewProgressTracker()
	defer tracker.Close()

	ch := tracker.Subscribe()

	go tracker.EmitPhaseCompleted("Preflight", 1, 8)

	select {
	case update := <-ch:
		if update.Phase != "Preflight" {
			t.Errorf("expected phase 'Preflight', got %s", update.Phase)
		}
		if update.Message != "Preflight completed" {
			t.Errorf("expected message 'Preflight completed', got %s", update.Message)
		}
		if update.ProgressPercent != 12 { // 1 * 100 / 8 = 12%
			t.Errorf("expected progress 12%%, got %d%%", update.ProgressPercent)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for phase completed event")
	}
}

func TestProgressTracker_EmitPhaseError(t *testing.T) {
	tracker := NewProgressTracker()
	defer tracker.Close()

	ch := tracker.Subscribe()

	go tracker.EmitPhaseError("Preflight", 1, 8, "Something went wrong")

	select {
	case update := <-ch:
		if update.Phase != "Preflight" {
			t.Errorf("expected phase 'Preflight', got %s", update.Phase)
		}
		if update.Message != "Something went wrong" {
			t.Errorf("expected error message, got %s", update.Message)
		}
		if !update.IsError {
			t.Error("expected IsError to be true")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for phase error event")
	}
}

// Note: Unsubscribe test is not practical since Subscribe returns a read-only channel
// and the internal channel is not directly accessible to the caller.

func TestProgressTracker_TimestampSet(t *testing.T) {
	tracker := NewProgressTracker()
	defer tracker.Close()

	ch := tracker.Subscribe()

	update := &dto.ProgressUpdate{
		Phase:   "Test",
		Message: "Test message",
		// Timestamp is zero
	}

	go tracker.Emit(update)

	select {
	case received := <-ch:
		if received.Timestamp.IsZero() {
			t.Error("expected timestamp to be set")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for progress update")
	}
}
