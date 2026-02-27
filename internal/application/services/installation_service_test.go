package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/handlers"
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func newMockResponse(ctrl *gomock.Controller, statusCode int, body []byte) ports.Response {
	resp := mocks.NewMockResponse(ctrl)
	resp.EXPECT().StatusCode().Return(statusCode).AnyTimes()
	resp.EXPECT().Body().Return(body).AnyTimes()
	resp.EXPECT().Close().Return(nil).AnyTimes()
	return resp
}

func createTestService(ctrl *gomock.Controller) *InstallationService {
	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockRepo := mocks.NewMockInstallationRepository(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("HOOKS=(base)\n"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Chmod(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChrootWithStdin(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	bootstrapHandler := handlers.NewBootstrapHandler(mockFS, mockHTTP, mockLogger, "https://github.com/bnema/archup", "https://raw.githubusercontent.com/bnema/archup/dev", "dev")
	preflightHandler := handlers.NewPreflightHandler(mockFS, mockExec, mockLogger)
	partitionHandler := handlers.NewPartitionHandler(mockExec, mockLogger)
	baseHandler := handlers.NewInstallBaseHandler(mockFS, mockExec, mockChrExec, mockLogger)
	configHandler := handlers.NewConfigureSystemHandler(mockFS, mockChrExec, mockLogger)
	bootloaderHandler := handlers.NewBootloaderHandler(mockFS, mockExec, mockChrExec, mockLogger)
	reposHandler := handlers.NewReposHandler(mockFS, mockChrExec, mockLogger)
	postInstallHandler := handlers.NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	return NewInstallationService(
		mockRepo,
		mockLogger,
		bootstrapHandler,
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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

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
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockRepo := mocks.NewMockInstallationRepository(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("HOOKS=(base)\n"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Chmod(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChrootWithStdin(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	bootstrapHandler := handlers.NewBootstrapHandler(mockFS, mockHTTP, mockLogger, "https://github.com/bnema/archup", "https://raw.githubusercontent.com/bnema/archup/dev", "dev")
	preflightHandler := handlers.NewPreflightHandler(mockFS, mockExec, mockLogger)
	partitionHandler := handlers.NewPartitionHandler(mockExec, mockLogger)
	baseHandler := handlers.NewInstallBaseHandler(mockFS, mockExec, mockChrExec, mockLogger)
	configHandler := handlers.NewConfigureSystemHandler(mockFS, mockChrExec, mockLogger)
	bootloaderHandler := handlers.NewBootloaderHandler(mockFS, mockExec, mockChrExec, mockLogger)
	reposHandler := handlers.NewReposHandler(mockFS, mockChrExec, mockLogger)
	postInstallHandler := handlers.NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	service := NewInstallationService(mockRepo, mockLogger, bootstrapHandler, preflightHandler, partitionHandler, baseHandler, configHandler, bootloaderHandler, reposHandler, postInstallHandler)
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	status := service.GetStatus()

	if status.State != "NotStarted" {
		t.Errorf("expected state NotStarted, got %s", status.State)
	}
}

func TestInstallationService_GetStatus_Started(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := createTestService(ctrl)
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ctx := context.Background()
	if err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ctx := context.Background()
	if err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ctx := context.Background()
	if err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cmd := commands.PartitionDiskCommand{
		TargetDisk:     "/dev/sda",
		RootSizeGB:     50,
		BootSizeGB:     1,
		EncryptionType: disk.EncryptionTypeNone,
		FilesystemType: disk.FilesystemExt4,
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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ctx := context.Background()
	if err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ctx := context.Background()
	if err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ctx := context.Background()
	if err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 5,
		Branding:       "ArchUp",
		KernelVariant:  packages.KernelStable,
		RootPartition:  "/dev/sda2",
		EncryptionType: disk.EncryptionTypeNone,
		EFIPartition:   "/dev/sda1",
		TargetDisk:     "/dev/sda",
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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ctx := context.Background()
	if err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ctx := context.Background()
	if err := service.Start(ctx, "myarch", "testuser", "/dev/sda", "none"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

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
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}()

	ch := service.SubscribeProgress()

	if ch == nil {
		t.Fatal("expected non-nil progress channel")
	}
}
