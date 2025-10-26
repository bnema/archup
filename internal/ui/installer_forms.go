package ui

import (
	"fmt"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/system"
	"github.com/bnema/archup/internal/ui/components"
	"github.com/bnema/archup/internal/validation"
	"github.com/charmbracelet/huh"
)

// CreatePreflightForm creates the initial configuration form with all user identity fields
func CreatePreflightForm(cfg *config.Config) *huh.Form {
	fb := components.NewFormBuilder(false)

	// Auto-detect timezone from API
	detectedTimezone := system.DetectTimezone()
	switch {
	case detectedTimezone != "":
		cfg.Timezone = detectedTimezone
	case cfg.Timezone == "":
		cfg.Timezone = config.TimezoneDefault
	}

	return fb.CreateForm(
		// System configuration
		huh.NewGroup(
			fb.TextInput("Hostname", "> ", &cfg.Hostname, validation.ValidateHostname),
			fb.TextInput("Username", "> ", &cfg.Username, validation.ValidateUsername),
		).Title(SectionUserIdentity).Description(SectionUserIdentity),

		// Passwords
		huh.NewGroup(
			fb.PasswordInput("User Password", "> ", &cfg.UserPassword, validation.ValidatePassword),
			fb.PasswordInput("Root Password", "> ", &cfg.RootPassword, validation.ValidatePassword),
		),

		// Email (optional)
		huh.NewGroup(
			fb.TextInput("Email (optional)", "> ", &cfg.Email, nil).
				Description(DescEmail),
		),

		// Localization
		huh.NewGroup(
			fb.TextInput("Timezone", "> ", &cfg.Timezone, validation.ValidateTimezone).
				Description(DescTimezone),
			fb.TextInput("Locale", "> ", &cfg.Locale, validation.ValidateLocale).
				Description(DescLocale),
			fb.TextInput("Keymap", "> ", &cfg.Keymap, nil).
				Description(DescKeymap),
		),
	)
}

// CreateDiskSelectionForm creates disk selection form with detailed disk information
func CreateDiskSelectionForm(cfg *config.Config) *huh.Form {
	fb := components.NewFormBuilder(false)

	// Detect available disks
	disks, err := system.ListDisks()

	var diskOptions []string
	switch {
	case err != nil:
		// Fallback to empty list if detection fails
		diskOptions = []string{"No disks detected"}
	default:
		// Format disk options with detailed info: "/dev/sda (500GB) Samsung SSD [S3Z9NB0K123456]"
		for _, disk := range disks {
			label := disk.Path + " (" + disk.Size + ")"

			// Add model if available
			switch {
			case disk.Model != "":
				label += " " + disk.Model
			}

			// Add serial if available
			switch {
			case disk.Serial != "":
				label += " [" + disk.Serial + "]"
			}

			// Add vendor if available
			switch {
			case disk.Vendor != "" && disk.Model == "":
				label += " " + disk.Vendor
			}

			diskOptions = append(diskOptions, label)
		}
	}

	return fb.CreateForm(
		huh.NewGroup(
			fb.Select(SectionDiskSelection, diskOptions, &cfg.TargetDisk).
				Description(WarnDiskErase),
		),
	)
}

// CreateOptionsForm creates installation options form with CPU detection and descriptions
func CreateOptionsForm(cfg *config.Config) *huh.Form {
	fb := components.NewFormBuilder(false)

	// Auto-detect CPU information
	cpuInfo, err := system.DetectCPUInfo()
	switch {
	case err == nil:
		cfg.CPUVendor = string(cpuInfo.Vendor)
		cfg.Microcode = cpuInfo.Microcode
	}

	// Kernel selection (always shown)
	kernelGroup := huh.NewGroup(
		fb.SelectWithOptions("Kernel", components.CreateKernelOptions(), &cfg.KernelChoice),
	).Title(SectionKernel)

	// Encryption options
	encryptionOptions := []huh.Option[string]{
		huh.NewOption("No encryption", config.EncryptionNone),
		huh.NewOption("LUKS encryption", config.EncryptionLUKS),
	}

	encryptionGroup := huh.NewGroup(
		fb.SelectWithOptions("Encryption", encryptionOptions, &cfg.EncryptionType),
	).Title(SectionDiskEncryption).Description(DescEncryption)

	// Repository options
	repoGroup := huh.NewGroup(
		fb.Confirm("Enable multilib repository?", "Yes", "No", &cfg.EnableMultilib).
			Description(DescMultilib),
		fb.Confirm("Enable Chaotic-AUR?", "Yes", "No", &cfg.EnableChaotic).
			Description(DescChaoticAUR),
	)

	// AUR helper options
	aurOptions := []huh.Option[string]{
		huh.NewOption("None", ""),
		components.CreateOption("Paru", DescAURHelperParu, config.AURHelperParu),
		components.CreateOption("Yay", DescAURHelperYay, config.AURHelperYay),
	}

	aurGroup := huh.NewGroup(
		fb.SelectWithOptions("AUR Helper", aurOptions, &cfg.AURHelper).
			Description(DescAURHelper),
	)

	// Build form groups list
	groups := []*huh.Group{
		kernelGroup,
		encryptionGroup,
		repoGroup,
		aurGroup,
	}

	// Add AMD P-State tuning group (conditional - only for AMD CPUs)
	switch {
	case cpuInfo != nil && cpuInfo.Vendor == system.CPUVendorAMD:
		switch {
		case len(cpuInfo.AMDPStateModes) == 1:
			// Auto-select if only one mode available
			cfg.AMDPState = string(cpuInfo.AMDPStateModes[0])
		case len(cpuInfo.AMDPStateModes) > 1:
			// Show selection if multiple modes available
			modeStrings := make([]string, len(cpuInfo.AMDPStateModes))
			for i, mode := range cpuInfo.AMDPStateModes {
				modeStrings[i] = string(mode)
			}

			amdGroup := huh.NewGroup(
				fb.SelectWithOptions("AMD P-State Mode",
					components.CreateAMDPStateOptions(modeStrings),
					&cfg.AMDPState),
			).Title(SectionAMDTuning).Description(DescAMDPState)

			// Insert AMD group after kernel group
			groups = append(groups[:1], append([]*huh.Group{amdGroup}, groups[1:]...)...)
		}
	}

	return fb.CreateForm(groups...)
}

// CreateEncryptionPasswordForm creates encryption password form
// User can leave empty to use same password as user account
func CreateEncryptionPasswordForm(cfg *config.Config) *huh.Form {
	fb := components.NewFormBuilder(false)

	return fb.CreateForm(
		huh.NewGroup(
			fb.PasswordInput("Encryption Password", "> ", &cfg.EncryptPassword, nil).
				Description(DescEncryptionPassword),
		).Title(SectionDiskEncryption).Description(DescEncryption),
	)
}

// FormatDiskOption formats a disk into a user-friendly option string (DRY helper)
func FormatDiskOption(disk system.Disk) string {
	label := fmt.Sprintf("%s (%s)", disk.Path, disk.Size)

	switch {
	case disk.Model != "":
		label += " " + disk.Model
	}

	switch {
	case disk.Serial != "":
		label += " [" + disk.Serial + "]"
	}

	switch {
	case disk.Vendor != "" && disk.Model == "":
		label += " " + disk.Vendor
	}

	return label
}
