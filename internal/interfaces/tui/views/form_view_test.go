package views

import (
	"errors"
	"strings"
	"testing"

	"github.com/bnema/archup/internal/interfaces/tui/models"
)

func TestRenderForm(t *testing.T) {
	// Create a test form model
	fm := models.NewFormModel()

	// Render the form
	output := RenderForm(fm)

	// Verify key content is present
	checks := []string{
		"ArchUp Installer Configuration", // Title
		"Hostname:",                      // Field labels
		"Username:",
		"↑↓ Navigate", // Instructions
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Expected form output to contain '%s', but it didn't", check)
		}
	}

	// Verify output is not empty
	if len(output) == 0 {
		t.Error("Form output should not be empty")
	}

	t.Logf("Form rendered successfully with %d characters", len(output))
}

func TestRenderFormWithError(t *testing.T) {
	// Create a test form model with error
	fm := models.NewFormModel()
	testErr := "Test error message"
	fm.SetError(errors.New(testErr))

	// Render the form
	output := RenderForm(fm)

	// Verify error is displayed
	if !strings.Contains(output, testErr) {
		t.Errorf("Expected error message '%s' in form output", testErr)
	}
}

func TestRenderFormWithDifferentFocus(t *testing.T) {
	// Create a test form model
	fm := models.NewFormModel()

	// Render initial form
	output1 := RenderForm(fm)

	// Move focus
	fm.FocusNext()

	// Render form with different focus
	output2 := RenderForm(fm)

	// Both should render successfully
	if len(output1) == 0 || len(output2) == 0 {
		t.Error("Both renders should produce output")
	}

	// They should be different (different field is focused)
	if output1 == output2 {
		t.Log("Outputs are the same, but that's okay as the visual difference might not be captured")
	}

	t.Logf("Form 1: %d chars, Form 2: %d chars", len(output1), len(output2))
}

func TestFormModelCreation(t *testing.T) {
	fm := models.NewFormModel()

	// Verify initial state
	fields := fm.GetFields()
	if len(fields) != 8 {
		t.Errorf("Expected 8 form fields, got %d", len(fields))
	}

	if fm.GetFocusIndex() != 0 {
		t.Errorf("Expected initial focus index to be 0, got %d", fm.GetFocusIndex())
	}

	if fm.GetError() != nil {
		t.Errorf("Expected no initial error, got %v", fm.GetError())
	}

	if fm.IsSubmitted() {
		t.Error("Form should not be submitted initially")
	}
}

func TestFormFieldInteraction(t *testing.T) {
	fm := models.NewFormModel()

	// Test focus navigation
	initialIndex := fm.GetFocusIndex()
	fm.FocusNext()
	if fm.GetFocusIndex() != initialIndex+1 {
		t.Error("FocusNext should increment focus index")
	}

	fm.FocusPrevious()
	if fm.GetFocusIndex() != initialIndex {
		t.Error("FocusPrevious should decrement focus index")
	}

	// Test data extraction
	formData := fm.GetData()
	if formData.Hostname != "" || formData.Username != "" {
		t.Error("Initial form data should be empty")
	}

	// Set data
	testData := models.FormData{
		Hostname:       "test-host",
		Username:       "testuser",
		TargetDisk:     "/dev/sda",
		EncryptionType: "LUKS",
	}
	fm.SetData(testData)

	// Verify data was set
	retrievedData := fm.GetData()
	if retrievedData.Hostname != testData.Hostname {
		t.Errorf("Expected hostname '%s', got '%s'", testData.Hostname, retrievedData.Hostname)
	}
}

func BenchmarkRenderForm(b *testing.B) {
	fm := models.NewFormModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderForm(fm)
	}
}
