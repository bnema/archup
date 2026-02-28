package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports"
)

// BootloaderHandler handles bootloader installation
type BootloaderHandler struct {
	fs      ports.FileSystem
	cmdExec ports.CommandExecutor
	chrExec ports.ChrootExecutor
	logger  ports.Logger
}

// NewBootloaderHandler creates a new bootloader handler
func NewBootloaderHandler(fs ports.FileSystem, cmdExec ports.CommandExecutor, chrExec ports.ChrootExecutor, logger ports.Logger) *BootloaderHandler {
	return &BootloaderHandler{
		fs:      fs,
		cmdExec: cmdExec,
		chrExec: chrExec,
		logger:  logger,
	}
}

// Handle installs the bootloader
func (h *BootloaderHandler) Handle(ctx context.Context, cmd commands.InstallBootloaderCommand) (*dto.BootloaderResult, error) {
	h.logger.Info("Starting bootloader installation", "type", cmd.BootloaderType)

	h.logger.Info("Validating bootloader configuration...")
	result := &dto.BootloaderResult{
		Success:        false,
		BootloaderType: "",
		Timeout:        cmd.TimeoutSeconds,
		ErrorDetail:    "",
	}

	// Create bootloader domain object
	bl, err := bootloader.NewBootloader(cmd.BootloaderType, cmd.TimeoutSeconds, cmd.Branding)
	if err != nil {
		h.logger.Error("Invalid bootloader configuration", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid bootloader configuration: %v", err)
		return result, err
	}

	h.logger.Info("Bootloader configuration validated", "type", bl.Type())

	kernel, err := packages.NewKernel(cmd.KernelVariant)
	if err != nil {
		h.logger.Error("Invalid kernel variant", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid kernel variant: %v", err)
		return result, err
	}

	if err := h.configureMkinitcpio(ctx, cmd.MountPoint, cmd.EncryptionType, cmd.GPUVendor); err != nil {
		result.ErrorDetail = err.Error()
		return result, err
	}

	if err := h.installLimine(ctx, cmd.MountPoint); err != nil {
		result.ErrorDetail = err.Error()
		return result, err
	}

	if err := h.configureLimine(ctx, cmd, kernel.PackageName()); err != nil {
		result.ErrorDetail = err.Error()
		return result, err
	}

	if err := h.createBootEntry(ctx, cmd.TargetDisk, cmd.EFIPartition, cmd.MountPoint); err != nil {
		result.ErrorDetail = err.Error()
		return result, err
	}

	result.Success = true
	result.BootloaderType = bl.Type().String()
	result.Timeout = bl.Timeout()

	h.logger.Info("Bootloader installation completed successfully")
	return result, nil
}

func (h *BootloaderHandler) configureMkinitcpio(ctx context.Context, mountPoint string, encType disk.EncryptionType, gpuVendor string) error {
	confPath := filepath.Join(mountPoint, "etc", "mkinitcpio.conf")
	content, err := h.fs.ReadFile(confPath)
	if err != nil {
		h.logger.Error("Failed to read mkinitcpio.conf", "error", err)
		return fmt.Errorf("failed to read mkinitcpio.conf: %w", err)
	}

	hooks := config.MkinitcpioHooksPlymouth
	if encType != disk.EncryptionTypeNone {
		hooks = config.MkinitcpioHooksEncrypted
	}

	updated := replaceHooksLine(string(content), hooks)

	// Set the KMS module for early framebuffer so Plymouth loads cleanly.
	// Without this, the kms hook has nothing to load and Plymouth appears late.
	kmsModule := kmsModuleForGPU(gpuVendor)
	updated = replaceModulesLine(updated, kmsModule)

	if err := h.fs.WriteFile(confPath, []byte(updated), 0644); err != nil {
		h.logger.Error("Failed to write mkinitcpio.conf", "error", err)
		return fmt.Errorf("failed to write mkinitcpio.conf: %w", err)
	}

	if _, err := h.chrExec.ExecuteInChroot(ctx, mountPoint, "mkinitcpio", "-P"); err != nil {
		h.logger.Error("Failed to regenerate initramfs", "error", err)
		return fmt.Errorf("failed to regenerate initramfs: %w", err)
	}

	return nil
}

// kmsModuleForGPU returns the kernel module name required for early KMS on the given GPU vendor.
// Returns an empty string for NVIDIA (uses proprietary driver, no early KMS) or unknown GPUs.
func kmsModuleForGPU(vendor string) string {
	switch vendor {
	case "amd":
		return "amdgpu"
	case "intel":
		return "i915"
	default:
		return ""
	}
}

// replaceModulesLine replaces the MODULES=(...) line in mkinitcpio.conf.
// If module is empty, MODULES=() is left unchanged (or set to empty).
func replaceModulesLine(content, module string) string {
	re := regexp.MustCompile(`(?m)^MODULES=\(.*\)$`)
	if module == "" {
		return re.ReplaceAllString(content, "MODULES=()")
	}
	return re.ReplaceAllString(content, fmt.Sprintf("MODULES=(%s)", module))
}

func (h *BootloaderHandler) installLimine(ctx context.Context, mountPoint string) error {
	limineDir := filepath.Join(mountPoint, "boot", "EFI", "limine")
	if err := h.fs.MkdirAll(limineDir, 0755); err != nil {
		h.logger.Error("Failed to create Limine directory", "error", err)
		return fmt.Errorf("failed to create Limine directory: %w", err)
	}

	src := filepath.Join(mountPoint, "usr", "share", "limine", "BOOTX64.EFI")
	dst := filepath.Join(limineDir, "BOOTX64.EFI")
	if _, err := h.cmdExec.Execute(ctx, "cp", src, dst); err != nil {
		h.logger.Error("Failed to copy Limine EFI", "error", err)
		return fmt.Errorf("failed to copy Limine EFI: %w", err)
	}

	// Also copy to default/fallback boot path for UEFI firmwares that don't honor boot entries
	fallbackDir := filepath.Join(mountPoint, "boot", "EFI", "BOOT")
	if err := h.fs.MkdirAll(fallbackDir, 0755); err != nil {
		h.logger.Error("Failed to create fallback EFI directory", "error", err)
		return fmt.Errorf("failed to create fallback EFI directory: %w", err)
	}
	fallbackDst := filepath.Join(fallbackDir, "BOOTX64.EFI")
	if _, err := h.cmdExec.Execute(ctx, "cp", src, fallbackDst); err != nil {
		h.logger.Error("Failed to copy Limine EFI to fallback path", "error", err)
		return fmt.Errorf("failed to copy Limine EFI to fallback path: %w", err)
	}

	return nil
}

func (h *BootloaderHandler) configureLimine(ctx context.Context, cmd commands.InstallBootloaderCommand, kernelName string) error {
	rootUUIDBytes, err := h.cmdExec.Execute(ctx, "blkid", "-s", "UUID", "-o", "value", cmd.RootPartition)
	if err != nil {
		h.logger.Error("Failed to get root UUID", "error", err)
		return fmt.Errorf("failed to get root UUID: %w", err)
	}

	rootUUID := strings.TrimSpace(string(rootUUIDBytes))
	var kernelParams string
	if cmd.EncryptionType != disk.EncryptionTypeNone {
		kernelParams = fmt.Sprintf("cryptdevice=UUID=%s:cryptroot root=/dev/mapper/cryptroot rootflags=subvol=@ rw", rootUUID)
	} else {
		kernelParams = fmt.Sprintf("root=UUID=%s rootflags=subvol=@ rw", rootUUID)
	}

	kernelParams = strings.TrimSpace(fmt.Sprintf("%s %s", kernelParams, config.KernelParamsQuiet))
	if extra := strings.TrimSpace(cmd.KernelParamsExtra); extra != "" {
		kernelParams = strings.TrimSpace(kernelParams + " " + extra)
	}

	templatePath, err := h.resolveLimineTemplate()
	if err != nil {
		h.logger.Error("Failed to locate Limine template", "error", err)
		return fmt.Errorf("failed to locate Limine template: %w", err)
	}

	templateBytes, err := h.fs.ReadFile(templatePath)
	if err != nil {
		h.logger.Error("Failed to read Limine template", "error", err)
		return fmt.Errorf("failed to read Limine template: %w", err)
	}

	// Read machine-id from the installed system for limine-snapper-sync identification
	machineIDBytes, err := h.fs.ReadFile(filepath.Join(cmd.MountPoint, "etc", "machine-id"))
	machineID := strings.TrimSpace(string(machineIDBytes))
	if err != nil || machineID == "" {
		machineID = "unknown"
	}

	limineConfig := string(templateBytes)
	limineConfig = strings.ReplaceAll(limineConfig, "{{TIMEOUT}}", fmt.Sprintf("%d", cmd.TimeoutSeconds))
	limineConfig = strings.ReplaceAll(limineConfig, "{{BRANDING}}", cmd.Branding)
	limineConfig = strings.ReplaceAll(limineConfig, "{{COLOR}}", config.LimineColor)
	limineConfig = strings.ReplaceAll(limineConfig, "{{KERNEL}}", kernelName)
	limineConfig = strings.ReplaceAll(limineConfig, "{{KERNEL_PARAMS}}", kernelParams)
	limineConfig = strings.ReplaceAll(limineConfig, "{{MACHINE_ID}}", machineID)

	limineConfigPath := filepath.Join(cmd.MountPoint, "boot", "limine.conf")
	if err := h.fs.WriteFile(limineConfigPath, []byte(limineConfig), 0644); err != nil {
		h.logger.Error("Failed to write Limine config", "error", err)
		return fmt.Errorf("failed to write Limine config: %w", err)
	}

	return nil
}

func (h *BootloaderHandler) createBootEntry(ctx context.Context, targetDisk, efiPartition, mountPoint string) error {
	partNum := extractPartitionNumber(efiPartition)
	if partNum == "" {
		return fmt.Errorf("failed to determine EFI partition number from %s", efiPartition)
	}

	if _, err := h.chrExec.ExecuteInChroot(ctx, mountPoint, "efibootmgr", "--create", "--disk", targetDisk, "--part", partNum, "--label", config.UEFIBootLabel, "--loader", config.UEFIBootLoader, "--unicode"); err != nil {
		h.logger.Error("Failed to create EFI boot entry", "error", err)
		return fmt.Errorf("failed to create EFI boot entry: %w", err)
	}

	return nil
}

func replaceHooksLine(content string, hooks string) string {
	re := regexp.MustCompile(`(?m)^HOOKS=.*$`)
	return re.ReplaceAllString(content, hooks)
}

func extractPartitionNumber(partition string) string {
	re := regexp.MustCompile(`[0-9]+$`)
	return re.FindString(partition)
}

func (h *BootloaderHandler) resolveLimineTemplate() (string, error) {
	candidates := []string{
		filepath.Join(config.DefaultInstallDir, "configs", "limine.conf.template"),
		filepath.Join("install", "configs", "limine.conf.template"),
	}

	for _, candidate := range candidates {
		exists, err := h.fs.Exists(candidate)
		if err == nil && exists {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("limine.conf.template not found")
}
