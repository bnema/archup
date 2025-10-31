package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestReposHandler_Handle_Minimal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewReposHandler(mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:      "/mnt",
		EnableMultilib:  false,
		EnableChaotic:   false,
		AURHelper:       packages.AURHelperParu,
		AdditionalRepos: []string{},
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if result.Multilib {
		t.Error("expected multilib to be disabled")
	}

	if result.Chaotic {
		t.Error("expected chaotic to be disabled")
	}

	if result.AURHelper != "paru" {
		t.Errorf("expected paru helper, got %s", result.AURHelper)
	}
}

func TestReposHandler_Handle_AllEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewReposHandler(mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:      "/mnt",
		EnableMultilib:  true,
		EnableChaotic:   true,
		AURHelper:       packages.AURHelperYay,
		AdditionalRepos: []string{},
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if !result.Multilib {
		t.Error("expected multilib to be enabled")
	}

	if !result.Chaotic {
		t.Error("expected chaotic to be enabled")
	}

	if result.AURHelper != "yay" {
		t.Errorf("expected yay helper, got %s", result.AURHelper)
	}
}

func TestReposHandler_Handle_WithAdditionalRepos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewReposHandler(mockChrExec, mockLogger)

	additionalRepos := []string{
		"https://example.com/archup",
		"https://custom.repo/packages",
	}

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:      "/mnt",
		EnableMultilib:  false,
		EnableChaotic:   false,
		AURHelper:       packages.AURHelperParu,
		AdditionalRepos: additionalRepos,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}
