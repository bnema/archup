package user

import (
	"testing"
)

func TestNewUser_Valid(t *testing.T) {
	user, err := NewUser("testuser", "/bin/bash")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Username() != "testuser" {
		t.Errorf("expected username testuser")
	}

	if user.Shell() != "/bin/bash" {
		t.Errorf("expected shell /bin/bash")
	}

	if user.Home() != "/home/testuser" {
		t.Errorf("expected home /home/testuser")
	}
}

func TestNewUser_InvalidUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		shouldErr bool
	}{
		{"empty username", "", true},
		{"too short", "ab", true},
		{"too long", "a" + string(make([]byte, 100)), true},
		{"starts with digit", "1user", true},
		{"uppercase", "User", true},
		{"contains space", "test user", true},
		{"valid simple", "testuser", false},
		{"with underscore", "test_user", false},
		{"with hyphen", "test-user", false},
		{"with digit", "user123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser(tt.username, "/bin/bash")
			if (err != nil) != tt.shouldErr {
				t.Errorf("expected error=%v, got %v", tt.shouldErr, err)
			}
		})
	}
}

func TestNewUser_InvalidShell(t *testing.T) {
	tests := []struct {
		name      string
		shell     string
		shouldErr bool
	}{
		{"empty shell", "", true},
		{"relative path", "bash", true},
		{"valid bash", "/bin/bash", false},
		{"valid zsh", "/bin/zsh", false},
		{"valid sh", "/bin/sh", false},
		{"valid fish", "/bin/fish", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser("testuser", tt.shell)
			if (err != nil) != tt.shouldErr {
				t.Errorf("expected error=%v, got %v", tt.shouldErr, err)
			}
		})
	}
}

func TestUser_AddGroup(t *testing.T) {
	user, _ := NewUser("testuser", "/bin/bash")

	err := user.AddGroup("wheel")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !user.HasGroup("wheel") {
		t.Error("expected user to be in wheel group")
	}
}

func TestUser_AddGroup_Duplicate(t *testing.T) {
	user, _ := NewUser("testuser", "/bin/bash")

	if err := user.AddGroup("wheel"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err := user.AddGroup("wheel")

	if err == nil {
		t.Error("expected error for duplicate group")
	}
}

func TestUser_AddGroup_Empty(t *testing.T) {
	user, _ := NewUser("testuser", "/bin/bash")

	err := user.AddGroup("")

	if err == nil {
		t.Error("expected error for empty group")
	}
}

func TestUser_HasGroup(t *testing.T) {
	user, _ := NewUser("testuser", "/bin/bash")

	if user.HasGroup("wheel") {
		t.Error("expected user not in wheel group")
	}

	if err := user.AddGroup("wheel"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !user.HasGroup("wheel") {
		t.Error("expected user in wheel group")
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		shouldErr bool
	}{
		{"valid", "testuser", false},
		{"empty", "", true},
		{"short", "ab", true},
		{"long", "a" + string(make([]byte, 100)), true},
		{"starts digit", "1test", true},
		{"uppercase", "TestUser", true},
		{"special", "test@user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.shouldErr {
				t.Errorf("expected error=%v, got %v", tt.shouldErr, err)
			}
		})
	}
}

func TestNewCredentials_Valid(t *testing.T) {
	creds, err := NewCredentials("userpass123", "rootpass456")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !creds.HasPassword() {
		t.Error("expected to have password")
	}
}

func TestNewCredentials_WeakPassword(t *testing.T) {
	_, err := NewCredentials("wee", "rootpass456")

	if err == nil {
		t.Error("expected error for weak password")
	}
}

func TestNewCredentials_SamePassword(t *testing.T) {
	_, err := NewCredentials("samepass123", "samepass123")

	if err == nil {
		t.Error("expected error for same passwords")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		shouldErr bool
	}{
		{"empty", "", true},
		{"too short", "sho", true},
		{"valid", "validpass123", false},
		{"long valid", "a" + string(make([]byte, 100)) + "x", false},
		{"too long", "a" + string(make([]byte, 200)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.shouldErr {
				t.Errorf("expected error=%v, got %v", tt.shouldErr, err)
			}
		})
	}
}
