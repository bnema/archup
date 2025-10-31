package services

import (
	"sync"
	"time"

	"github.com/bnema/archup/internal/application/dto"
)

// ProgressTracker tracks and broadcasts installation progress
type ProgressTracker struct {
	subscribers []chan *dto.ProgressUpdate
	mu          sync.RWMutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		subscribers: make([]chan *dto.ProgressUpdate, 0),
	}
}

// Subscribe returns a channel for progress updates
func (pt *ProgressTracker) Subscribe() <-chan *dto.ProgressUpdate {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	ch := make(chan *dto.ProgressUpdate, 10)
	pt.subscribers = append(pt.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscriber
func (pt *ProgressTracker) Unsubscribe(ch chan *dto.ProgressUpdate) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for i, sub := range pt.subscribers {
		if sub == ch {
			pt.subscribers = append(pt.subscribers[:i], pt.subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

// Emit sends a progress update to all subscribers
func (pt *ProgressTracker) Emit(update *dto.ProgressUpdate) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if update.Timestamp.IsZero() {
		update.Timestamp = time.Now()
	}

	for _, ch := range pt.subscribers {
		select {
		case ch <- update:
		default:
			// Channel full, skip to avoid blocking
		}
	}
}

// EmitPhaseStarted sends a phase started event
func (pt *ProgressTracker) EmitPhaseStarted(phase string, phaseNumber int, totalPhases int) {
	pt.Emit(&dto.ProgressUpdate{
		Phase:           phase,
		PhaseNumber:     phaseNumber,
		TotalPhases:     totalPhases,
		ProgressPercent: (phaseNumber - 1) * 100 / totalPhases,
		Message:         "Starting " + phase,
		IsError:         false,
	})
}

// EmitPhaseCompleted sends a phase completed event
func (pt *ProgressTracker) EmitPhaseCompleted(phase string, phaseNumber int, totalPhases int) {
	pt.Emit(&dto.ProgressUpdate{
		Phase:           phase,
		PhaseNumber:     phaseNumber,
		TotalPhases:     totalPhases,
		ProgressPercent: phaseNumber * 100 / totalPhases,
		Message:         phase + " completed",
		IsError:         false,
	})
}

// EmitPhaseError sends a phase error event
func (pt *ProgressTracker) EmitPhaseError(phase string, phaseNumber int, totalPhases int, errorMsg string) {
	pt.Emit(&dto.ProgressUpdate{
		Phase:           phase,
		PhaseNumber:     phaseNumber,
		TotalPhases:     totalPhases,
		ProgressPercent: (phaseNumber - 1) * 100 / totalPhases,
		Message:         errorMsg,
		IsError:         true,
	})
}

// Close closes all subscriber channels
func (pt *ProgressTracker) Close() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for _, ch := range pt.subscribers {
		close(ch)
	}
	pt.subscribers = nil
}
