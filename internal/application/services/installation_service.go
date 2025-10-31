package services

import (
	"context"
	"errors"
	"time"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/application/handlers"
	"github.com/bnema/archup/internal/domain/installation"
	"github.com/bnema/archup/internal/domain/ports"
)

// InstallationService orchestrates the entire installation process
type InstallationService struct {
	// Domain
	installAgg *installation.Installation

	// Handlers
	preflightHandler       *handlers.PreflightHandler
	partitionHandler       *handlers.PartitionHandler
	baseHandler            *handlers.InstallBaseHandler
	configHandler          *handlers.ConfigureSystemHandler
	bootloaderHandler      *handlers.BootloaderHandler
	reposHandler           *handlers.ReposHandler
	postInstallHandler     *handlers.PostInstallHandler

	// Ports
	repo   ports.InstallationRepository
	logger ports.Logger

	// Progress tracking
	tracker *ProgressTracker

	// State
	startTime time.Time
}

// NewInstallationService creates a new installation service with all handlers
func NewInstallationService(
	repo ports.InstallationRepository,
	logger ports.Logger,
	preflightHandler *handlers.PreflightHandler,
	partitionHandler *handlers.PartitionHandler,
	baseHandler *handlers.InstallBaseHandler,
	configHandler *handlers.ConfigureSystemHandler,
	bootloaderHandler *handlers.BootloaderHandler,
	reposHandler *handlers.ReposHandler,
	postInstallHandler *handlers.PostInstallHandler,
) *InstallationService {
	return &InstallationService{
		repo:                   repo,
		logger:                 logger,
		preflightHandler:       preflightHandler,
		partitionHandler:       partitionHandler,
		baseHandler:            baseHandler,
		configHandler:          configHandler,
		bootloaderHandler:      bootloaderHandler,
		reposHandler:           reposHandler,
		postInstallHandler:     postInstallHandler,
		tracker:                NewProgressTracker(),
	}
}

// Start begins a new installation
func (s *InstallationService) Start(ctx context.Context, hostname string, username string, targetDisk string, encryptionType string) error {
	s.logger.Info("Initializing installation", "hostname", hostname, "username", username)

	// Create installation aggregate
	inst, err := installation.NewInstallation(hostname, username, targetDisk, encryptionType)
	if err != nil {
		s.logger.Error("Failed to create installation", "error", err)
		s.tracker.EmitPhaseError("initialization", 0, 8, err.Error())
		return err
	}

	s.installAgg = inst
	s.startTime = time.Now()

	// Start the installation
	if err := s.installAgg.Start(ctx); err != nil {
		s.logger.Error("Failed to start installation", "error", err)
		s.tracker.EmitPhaseError("initialization", 0, 8, err.Error())
		return err
	}

	return nil
}

// RunPreflight runs preflight checks
func (s *InstallationService) RunPreflight(ctx context.Context) (*dto.PreflightResult, error) {
	if s.installAgg == nil {
		return nil, errors.New("installation not started")
	}

	s.tracker.EmitPhaseStarted("Preflight Checks", 1, 8)
	defer s.tracker.EmitPhaseCompleted("Preflight Checks", 1, 8)

	result, err := s.preflightHandler.Handle(ctx, commands.PreflightCommand{})
	if err != nil {
		s.tracker.EmitPhaseError("Preflight Checks", 1, 8, err.Error())
		return nil, err
	}

	if !result.ChecksPassed {
		errMsg := "Preflight checks failed"
		if len(result.CriticalErrors) > 0 {
			errMsg = result.CriticalErrors[0]
		}
		s.tracker.EmitPhaseError("Preflight Checks", 1, 8, errMsg)
		return result, errors.New(errMsg)
	}

	return result, nil
}

// RunPartition runs disk partitioning
func (s *InstallationService) RunPartition(ctx context.Context, cmd commands.PartitionDiskCommand) (*dto.PartitionResult, error) {
	if s.installAgg == nil {
		return nil, errors.New("installation not started")
	}

	s.tracker.EmitPhaseStarted("Disk Partitioning", 2, 8)
	defer s.tracker.EmitPhaseCompleted("Disk Partitioning", 2, 8)

	result, err := s.partitionHandler.Handle(ctx, cmd)
	if err != nil {
		s.tracker.EmitPhaseError("Disk Partitioning", 2, 8, err.Error())
		return nil, err
	}

	if !result.Success {
		s.tracker.EmitPhaseError("Disk Partitioning", 2, 8, result.ErrorDetail)
		return result, errors.New(result.ErrorDetail)
	}

	return result, nil
}

// RunBaseInstall runs base system installation
func (s *InstallationService) RunBaseInstall(ctx context.Context, cmd commands.InstallBaseCommand) (*dto.InstallBaseResult, error) {
	if s.installAgg == nil {
		return nil, errors.New("installation not started")
	}

	s.tracker.EmitPhaseStarted("Base Installation", 3, 8)
	defer s.tracker.EmitPhaseCompleted("Base Installation", 3, 8)

	result, err := s.baseHandler.Handle(ctx, cmd)
	if err != nil {
		s.tracker.EmitPhaseError("Base Installation", 3, 8, err.Error())
		return nil, err
	}

	if !result.Success {
		s.tracker.EmitPhaseError("Base Installation", 3, 8, result.ErrorDetail)
		return result, errors.New(result.ErrorDetail)
	}

	return result, nil
}

// RunConfigSystem runs system configuration
func (s *InstallationService) RunConfigSystem(ctx context.Context, cmd commands.ConfigureSystemCommand) (*dto.ConfigureSystemResult, error) {
	if s.installAgg == nil {
		return nil, errors.New("installation not started")
	}

	s.tracker.EmitPhaseStarted("System Configuration", 4, 8)
	defer s.tracker.EmitPhaseCompleted("System Configuration", 4, 8)

	result, err := s.configHandler.Handle(ctx, cmd)
	if err != nil {
		s.tracker.EmitPhaseError("System Configuration", 4, 8, err.Error())
		return nil, err
	}

	if !result.Success {
		s.tracker.EmitPhaseError("System Configuration", 4, 8, result.ErrorDetail)
		return result, errors.New(result.ErrorDetail)
	}

	return result, nil
}

// RunBootloaderSetup runs bootloader setup
func (s *InstallationService) RunBootloaderSetup(ctx context.Context, cmd commands.InstallBootloaderCommand) (*dto.BootloaderResult, error) {
	if s.installAgg == nil {
		return nil, errors.New("installation not started")
	}

	s.tracker.EmitPhaseStarted("Bootloader Setup", 5, 8)
	defer s.tracker.EmitPhaseCompleted("Bootloader Setup", 5, 8)

	result, err := s.bootloaderHandler.Handle(ctx, cmd)
	if err != nil {
		s.tracker.EmitPhaseError("Bootloader Setup", 5, 8, err.Error())
		return nil, err
	}

	if !result.Success {
		s.tracker.EmitPhaseError("Bootloader Setup", 5, 8, result.ErrorDetail)
		return result, errors.New(result.ErrorDetail)
	}

	return result, nil
}

// RunRepositorySetup runs repository setup
func (s *InstallationService) RunRepositorySetup(ctx context.Context, cmd commands.SetupRepositoriesCommand) (*dto.RepositoriesResult, error) {
	if s.installAgg == nil {
		return nil, errors.New("installation not started")
	}

	s.tracker.EmitPhaseStarted("Repository Setup", 6, 8)
	defer s.tracker.EmitPhaseCompleted("Repository Setup", 6, 8)

	result, err := s.reposHandler.Handle(ctx, cmd)
	if err != nil {
		s.tracker.EmitPhaseError("Repository Setup", 6, 8, err.Error())
		return nil, err
	}

	if !result.Success {
		s.tracker.EmitPhaseError("Repository Setup", 6, 8, result.ErrorDetail)
		return result, errors.New(result.ErrorDetail)
	}

	return result, nil
}

// RunPostInstall runs post-installation tasks
func (s *InstallationService) RunPostInstall(ctx context.Context, cmd commands.PostInstallCommand) (*dto.PostInstallResult, error) {
	if s.installAgg == nil {
		return nil, errors.New("installation not started")
	}

	s.tracker.EmitPhaseStarted("Post-Installation", 7, 8)
	defer s.tracker.EmitPhaseCompleted("Post-Installation", 7, 8)

	result, err := s.postInstallHandler.Handle(ctx, cmd)
	if err != nil {
		s.tracker.EmitPhaseError("Post-Installation", 7, 8, err.Error())
		return nil, err
	}

	if !result.Success {
		s.tracker.EmitPhaseError("Post-Installation", 7, 8, result.ErrorDetail)
		return result, errors.New(result.ErrorDetail)
	}

	return result, nil
}

// Complete marks installation as complete
func (s *InstallationService) Complete(ctx context.Context) error {
	if s.installAgg == nil {
		return errors.New("installation not started")
	}

	duration := int(time.Since(s.startTime).Seconds())
	if err := s.installAgg.Complete(duration); err != nil {
		s.logger.Error("Failed to complete installation", "error", err)
		s.tracker.EmitPhaseError("Completion", 8, 8, err.Error())
		return err
	}

	s.tracker.EmitPhaseCompleted("Completion", 8, 8)
	s.logger.Info("Installation completed successfully", "duration", duration)

	return nil
}

// GetStatus returns the current installation status
func (s *InstallationService) GetStatus() *dto.InstallationStatus {
	if s.installAgg == nil {
		return &dto.InstallationStatus{
			State: "NotStarted",
		}
	}

	completedAt := s.installAgg.CompletedAt()
	startedAt := s.installAgg.StartedAt()

	return &dto.InstallationStatus{
		ID:             s.installAgg.ID(),
		State:          s.installAgg.State().String(),
		Hostname:       s.installAgg.Hostname(),
		Username:       s.installAgg.Username(),
		TargetDisk:     s.installAgg.TargetDisk(),
		EncryptionType: s.installAgg.EncryptionType(),
		Progress:       s.installAgg.ProgressPercentage(),
		StartedAt:      startedAt,
		CompletedAt:    completedAt,
		CurrentPhase:   s.installAgg.State().String(),
	}
}

// SubscribeProgress returns a channel for progress updates
func (s *InstallationService) SubscribeProgress() <-chan *dto.ProgressUpdate {
	return s.tracker.Subscribe()
}

// Close closes the service
func (s *InstallationService) Close() error {
	s.tracker.Close()
	return nil
}
