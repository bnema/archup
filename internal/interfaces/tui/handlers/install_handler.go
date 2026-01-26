package handlers

import (
	"strings"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/interfaces/tui/models"
	tea "github.com/charmbracelet/bubbletea"
)

// NOTE: This package uses AppContext interface to avoid importing parent tui package
// This breaks the circular dependency: tui → tui/handlers → tui
// Messages are typed by the caller (app.go), not here

// ProgressUpdateMsg wraps progress update for bubbletea
type ProgressUpdateMsg struct {
	Update *dto.ProgressUpdate
}

// InstallationErrorMsg wraps installation error for bubbletea
type InstallationErrorMsg struct {
	Err error
}

// InstallationCompleteMsg signals installation completion
type InstallationCompleteMsg struct {
	Duration int
}

// HandleProgressUpdate processes progress update messages
func HandleProgressUpdate(app AppContext, msg interface{}, progressModel *models.ProgressModelImpl) (*models.ProgressModelImpl, tea.Cmd) {
	if m, ok := msg.(ProgressUpdateMsg); ok {
		progressModel.UpdateProgress(m.Update)
	}
	return progressModel, nil
}

// HandleInstallationError processes installation error messages
func HandleInstallationError(app AppContext, msg interface{}, installModel *models.InstallationModelImpl) (*models.InstallationModelImpl, tea.Cmd) {
	if errMsg, ok := msg.(InstallationErrorMsg); ok {
		app.GetLogger().Error("Installation error", "error", errMsg.Err)
		installModel.SetError(errMsg.Err.Error())
	}
	return installModel, nil
}

// HandleInstallationComplete processes installation completion messages
func HandleInstallationComplete(app AppContext, msg interface{}, installModel *models.InstallationModelImpl) (*models.InstallationModelImpl, tea.Cmd) {
	app.GetLogger().Info("Installation completed successfully")
	installModel.SetComplete()
	// Update status with final timestamps for duration display
	if svc := app.GetInstallService(); svc != nil {
		installModel.SetStatus(svc.GetStatus())
	}
	return installModel, nil
}

// CreateInstallationCommand creates an installation command from form data
// Runs the full installation in a goroutine and sends progress via p.Send()
func CreateInstallationCommand(app AppContext, formData models.FormData) tea.Cmd {
	return func() tea.Msg {
		ctx := app.GetContext()
		svc := app.GetInstallService()
		logger := app.GetLogger()
		program := app.GetProgram()

		encryptionType := normalizeEncryptionType(formData.EncryptionType)

		// Start installation via service
		if err := svc.Start(ctx, formData.Hostname, formData.Username, formData.TargetDisk, encryptionType); err != nil {
			logger.Error("Failed to start installation", "error", err)
			return InstallationErrorMsg{Err: err}
		}

		// Subscribe to progress updates and forward to bubbletea
		progressChan := app.GetProgressTracker().Subscribe()
		go func() {
			for {
				select {
				case update, ok := <-progressChan:
					if !ok {
						return
					}
					if program != nil {
						program.Send(ProgressUpdateMsg{Update: update})
					}
				case <-ctx.Done():
					logger.Info("Installation context cancelled")
					return
				}
			}
		}()

		// Run installation phases in background goroutine
		go func() {
			// Phase 1: Preflight
			if _, err := svc.RunPreflight(ctx); err != nil {
				logger.Error("Preflight checks failed", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}

			// Phase 1.5: Bootstrap (clone repo / download files)
			if _, err := svc.RunBootstrap(ctx); err != nil {
				logger.Error("Bootstrap failed", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}

			// Phase 2: Partition disk with chosen encryption using user password
			partitionCmd := commands.PartitionDiskCommand{
				TargetDisk:         formData.TargetDisk,
				RootSizeGB:         0, // Use all available space
				BootSizeGB:         4, // 4GB for limine-snapper-sync
				EncryptionType:     parseEncryptionType(formData.EncryptionType),
				EncryptionPassword: formData.UserPassword, // Use account password for disk encryption
				FilesystemType:     disk.FilesystemBtrfs,
				WipeDisks:          true,
			}
			partitionResult, err := svc.RunPartition(ctx, partitionCmd)
			if err != nil {
				logger.Error("Partitioning failed", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}

			// Phase 3: Base install
			baseCmd := commands.InstallBaseCommand{
				TargetDisk:       formData.TargetDisk,
				MountPoint:       "/mnt",
				KernelVariant:    parseKernelVariant(formData.KernelVariant),
				IncludeMicrocode: formData.Microcode,
			}
			if _, err := svc.RunBaseInstall(ctx, baseCmd); err != nil {
				logger.Error("Base installation failed", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}

			// Phase 4: Configure system
			configCmd := commands.ConfigureSystemCommand{
				MountPoint:   "/mnt",
				Hostname:     formData.Hostname,
				Timezone:     formData.Timezone,
				Locale:       formData.Locale,
				Keymap:       formData.Keymap,
				Username:     formData.Username,
				UserShell:    "/bin/bash",
				UserPassword: formData.UserPassword,
				RootPassword: formData.RootPassword,
			}
			if _, err := svc.RunConfigSystem(ctx, configCmd); err != nil {
				logger.Error("System configuration failed", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}

			// Phase 5: Bootloader
			bootCmd := commands.InstallBootloaderCommand{
				MountPoint:        "/mnt",
				BootloaderType:    bootloader.BootloaderTypeLimine,
				TimeoutSeconds:    5,
				Branding:          "ArchUp",
				KernelVariant:     parseKernelVariant(formData.KernelVariant),
				RootPartition:     partitionResult.RootPartition,
				EncryptionType:    partitionCmd.EncryptionType,
				EFIPartition:      partitionResult.EFIPartition,
				TargetDisk:        formData.TargetDisk,
				KernelParamsExtra: formData.KernelParamsExtra,
			}
			if _, err := svc.RunBootloaderSetup(ctx, bootCmd); err != nil {
				logger.Error("Bootloader setup failed", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}

			// Phase 6: Repository setup
			repoCmd := commands.SetupRepositoriesCommand{
				MountPoint:     "/mnt",
				EnableMultilib: true,
				EnableChaotic:  true,
				AURHelper:      parseAURHelper(formData.AURHelper),
			}
			if _, err := svc.RunRepositorySetup(ctx, repoCmd); err != nil {
				logger.Error("Repository setup failed", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}

			// Phase 7: Post-install
			postCmd := commands.PostInstallCommand{
				MountPoint:         "/mnt",
				Username:           formData.Username,
				UserEmail:          formData.UserEmail,
				PlymouthTheme:      config.PlymouthThemeName,
				RunPostBootScripts: true,
			}
			if _, err := svc.RunPostInstall(ctx, postCmd); err != nil {
				logger.Error("Post-installation failed", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}
			// Complete installation
			if err := svc.Complete(ctx); err != nil {
				logger.Error("Failed to complete installation", "error", err)
				if program != nil {
					program.Send(InstallationErrorMsg{Err: err})
				}
				return
			}

			// Send completion message
			if program != nil {
				program.Send(InstallationCompleteMsg{Duration: 0})
			}
		}()

		return nil
	}
}

// parseKernelVariant converts string to KernelVariant
func parseKernelVariant(s string) packages.KernelVariant {
	switch s {
	case "linux-zen", "zen":
		return packages.KernelZen
	case "linux-lts", "lts":
		return packages.KernelLTS
	case "linux-hardened", "hardened":
		return packages.KernelHardened
	case "linux-cachyos", "cachyos":
		return packages.KernelCachyOS
	default:
		return packages.KernelStable
	}
}

// parseAURHelper converts string to AURHelper
func parseAURHelper(s string) packages.AURHelper {
	switch s {
	case "yay":
		return packages.AURHelperYay
	default:
		return packages.AURHelperParu
	}
}

func normalizeEncryptionType(value string) string {
	if value == "" {
		return "luks"
	}

	switch strings.ToLower(value) {
	case "none":
		return "none"
	case "luks", "luks+", "luks+standard":
		return "luks"
	case "luks-lvm", "luks+lvm":
		return "luks-lvm"
	default:
		return "luks"
	}
}

func parseEncryptionType(value string) disk.EncryptionType {
	switch normalizeEncryptionType(value) {
	case "none":
		return disk.EncryptionTypeNone
	case "luks-lvm":
		return disk.EncryptionTypeLUKSLVM
	default:
		return disk.EncryptionTypeLUKS
	}
}
