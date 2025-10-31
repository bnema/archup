package services

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/handlers"
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func createTestService(ctrl *gomock.Controller) *InstallationService {
	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockRepo := mocks.NewMockInstallationRepository(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).AnyTimes()

	preflightHandler := handlers.NewPreflightHandler(mockFS, mockExec, mockLogger)
	partitionHandler := handlers.NewPartitionHandler(mockExec, mockLogger)
	baseHandler := handlers.NewInstallBaseHandler(mockFS, mockExec, mockChrExec, mockLogger)
	configHandler := handlers.NewConfigureSystemHandler(mockChrExec, mockLogger)
	bootloaderHandler := handlers.NewBootloaderHandler(mockChrExec, mockLogger)
	reposHandler := handlers.NewReposHandler(mockChrExec, mockLogger)
	postInstallHandler := handlers.NewPostInstallHandler(mockChrExec, mockScriptExec, mockLogger)

	return NewInstallationService(
		mockRepo,
		mockLogger,
		preflightHandler,
		partitionHandler,
		baseHandler,
		configHandler,
		bootloaderHandler,
		reposHandler,
		postInstallHandler,
	)
}

func TestInstallationService_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if service.installAgg == nil {
		t.Fatal("expected installation to be created")
	}

	if service.installAgg.Hostname() != "myarch" {
		t.Errorf("expected hostname myarch, got %s", service.installAgg.Hostname())
	}
}

func TestInstallationService_Start_InvalidHostname(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockRepo := mocks.NewMockInstallationRepository(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	preflightHandler := handlers.NewPreflightHandler(mockFS, mockExec, mockLogger)
	partitionHandler := handlers.NewPartitionHandler(mockExec, mockLogger)
	baseHandler := handlers.NewInstallBaseHandler(mockFS, mockExec, mockChrExec, mockLogger)
	configHandler := handlers.NewConfigureSystemHandler(mockChrExec, mockLogger)
	bootloaderHandler := handlers.NewBootloaderHandler(mockChrExec, mockLogger)
	reposHandler := handlers.NewReposHandler(mockChrExec, mockLogger)
	postInstallHandler := handlers.NewPostInstallHandler(mockChrExec, mockScriptExec, mockLogger)

	service := NewInstallationService(mockRepo, mockLogger, preflightHandler, partitionHandler, baseHandler, configHandler, bootloaderHandler, reposHandler, postInstallHandler)
	defer service.Close()

	ctx := context.Background()
	err := service.Start(ctx, "", "testuser", "/dev/sda", "none")

	if err == nil {
		t.Error("expected error for invalid hostname")
	}

	if service.installAgg != nil {
		t.Error("expected installation not to be created")
	}
}

func TestInstallationService_GetStatus_NotStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	status := service.GetStatus()

	if status.State != "NotStarted" {
		t.Errorf("expected state NotStarted, got %s", status.State)
	}
}

func TestInstallationService_GetStatus_Started(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	status := service.GetStatus()

	if status.State != "PreflightChecks" {
		t.Errorf("expected state PreflightChecks, got %s", status.State)
	}

	if status.Hostname != "myarch" {
		t.Errorf("expected hostname myarch, got %s", status.Hostname)
	}

	if status.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", status.Username)
	}
}

func TestInstallationService_RunPreflight_NotRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	// We're probably not running as root, so preflight will fail
	result, _ := service.RunPreflight(ctx)

	// Either expect error or result indicating failure
	if result != nil && result.ChecksPassed {
		// If we're running as root, both err and checksPassed should be false
		t.Error("expected preflight to fail when not running as root")
	}
}

func TestInstallationService_RunPartition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	cmd := commands.PartitionDiskCommand{
		TargetDisk:         "/dev/sda",
		RootSizeGB:         50,
		BootSizeGB:         1,
		EncryptionType:     disk.EncryptionTypeNone,
		FilesystemType:     disk.FilesystemExt4,
	}

	result, err := service.RunPartition(ctx, cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.ErrorDetail)
	}
}

func TestInstallationService_RunBaseInstall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	cmd := commands.InstallBaseCommand{
		TargetDisk:       "/dev/sda",
		MountPoint:       "/mnt",
		Packages:         []string{},
		KernelVariant:    packages.KernelStable,
		IncludeMicrocode: false,
	}

	result, err := service.RunBaseInstall(ctx, cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

func TestInstallationService_RunConfigSystem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	cmd := commands.ConfigureSystemCommand{
		MountPoint:   "/mnt",
		Hostname:     "myarch",
		Timezone:     "UTC",
		Locale:       "en_US.UTF-8",
		Keymap:       "us",
		Username:     "testuser",
		UserShell:    "/bin/bash",
		UserPassword: "testpass123",
		RootPassword: "rootpass456",
	}

	result, err := service.RunConfigSystem(ctx, cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

func TestInstallationService_RunBootloaderSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 5,
		Branding:       "ArchUp",
	}

	result, err := service.RunBootloaderSetup(ctx, cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

func TestInstallationService_RunRepositorySetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:      "/mnt",
		EnableMultilib:  false,
		EnableChaotic:   false,
		AURHelper:       packages.AURHelperParu,
		AdditionalRepos: []string{},
	}

	result, err := service.RunRepositorySetup(ctx, cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

func TestInstallationService_RunPostInstall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	service.Start(ctx, "myarch", "testuser", "/dev/sda", "none")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "",
	}

	result, err := service.RunPostInstall(ctx, cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

func TestInstallationService_CompleteNotStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ctx := context.Background()
	// Don't start installation, try to complete it

	err := service.Complete(ctx)

	if err == nil {
		t.Error("expected error when completing unstarted installation")
	}
}

func TestInstallationService_SubscribeProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer service.Close()

	ch := service.SubscribeProgress()

	if ch == nil {
		t.Fatal("expected non-nil progress channel")
	}
}
