package ui

import (
	"fmt"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/phases"
	"github.com/bnema/archup/internal/system"
	"github.com/bnema/archup/internal/ui/components"
	"github.com/bnema/archup/internal/validation"
	"github.com/charmbracelet/huh"
)

// CreatePreflightForm creates the initial configuration form with all user identity fields
func CreatePreflightForm(cfg *config.Config, fb *components.FormBuilder, sys *system.System) *huh.Form {

	// Auto-detect timezone from API, fallback to default if detection fails
	detectedTimezone := sys.DetectTimezone()
	switch {
	case detectedTimezone != "":
		cfg.Timezone = detectedTimezone
	default:
		cfg.Timezone = config.TimezoneDefault
	}

	// Set default keymap if not already set
	switch {
	case cfg.Keymap == "":
		cfg.Keymap = config.KeymapDefault
	}

	return fb.CreateForm(
		// System configuration
		huh.NewGroup(
			fb.TextInput("Hostname", "> ", &cfg.Hostname, validation.ValidateHostname),
			fb.TextInput("Username", "> ", &cfg.Username, validation.ValidateUsername),
			fb.TextInput("Email (optional)", "> ", &cfg.Email, nil).
				Description(DescEmail),
		).Title(SectionUserIdentity),

		// Passwords
		huh.NewGroup(
			fb.PasswordInput("User Password", "> ", &cfg.UserPassword, validation.ValidatePassword),
			fb.PasswordInput("Confirm Password", "> ", &cfg.ConfirmPassword, validation.ValidatePasswordConfirmation(&cfg.UserPassword)),

			fb.Confirm("Use same password for disk encryption?", "Yes", "No", &cfg.UseSamePasswordForEncryption),
		),

		// Disk Encryption Password (only shown when NOT using same password)
		huh.NewGroup(
			fb.PasswordInput("Disk Encryption Password", "> ", &cfg.EncryptPassword, validation.ValidatePassword).
				Description("Enter a separate password to unlock your encrypted disk at boot"),
		).WithHideFunc(func() bool {
			return cfg.UseSamePasswordForEncryption // Hide when using same password
		}),

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
func CreateDiskSelectionForm(cfg *config.Config, fb *components.FormBuilder) *huh.Form {

	// Detect available disks
	disks, err := system.ListDisks()

	var diskOptions []huh.Option[string]
	switch {
	case err != nil:
		// Fallback to empty option if detection fails
		diskOptions = []huh.Option[string]{
			huh.NewOption("No disks detected", ""),
		}
	default:
		// Format disk options with detailed info: "/dev/sda (500GB) Samsung SSD [S3Z9NB0K123456]"
		// Use disk.Path as value and formatted string as display label
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

			diskOptions = append(diskOptions, huh.NewOption(label, disk.Path))
		}
	}

	return fb.CreateForm(
		huh.NewGroup(
			fb.SelectWithOptions(SectionDiskSelection, diskOptions, &cfg.TargetDisk).
				Description(WarnDiskErase),
		),
	)
}

// CreateOptionsForm creates installation options form with CPU detection and descriptions
func CreateOptionsForm(cfg *config.Config, fb *components.FormBuilder) *huh.Form {

	// Auto-detect CPU information
	cpuInfo, err := system.DetectCPUInfo()
	switch {
	case err == nil:
		cfg.CPUVendor = string(cpuInfo.Vendor)
		cfg.Microcode = cpuInfo.Microcode
	}

	// Page 1: Kernel and Encryption (system base configuration)
	encryptionOptions := []huh.Option[string]{
		huh.NewOption("No encryption", config.EncryptionNone),
		huh.NewOption("LUKS encryption", config.EncryptionLUKS),
	}

	systemBaseGroup := huh.NewGroup(
		fb.SelectWithOptions("Kernel", components.CreateKernelOptions(), &cfg.KernelChoice),
		fb.SelectWithOptions("Encryption", encryptionOptions, &cfg.EncryptionType).
			Description(DescEncryption),
	).Title(SectionKernel)

	// Page 2: Repository and package management options
	aurOptions := []huh.Option[string]{
		huh.NewOption("None", ""),
		components.CreateOption("Paru", DescAURHelperParu, phases.AURHelperParu),
		components.CreateOption("Yay", DescAURHelperYay, phases.AURHelperYay),
	}

	packageManagementGroup := huh.NewGroup(
		fb.Confirm("Enable multilib repository?", "Yes", "No", &cfg.EnableMultilib).
			Description(DescMultilib),
		fb.Confirm("Enable Chaotic-AUR?", "Yes", "No", &cfg.EnableChaotic).
			Description(DescChaoticAUR),
		fb.SelectWithOptions("AUR Helper", aurOptions, &cfg.AURHelper).
			Description(DescAURHelper),
	)

	// Build form groups list
	groups := []*huh.Group{
		systemBaseGroup,
		packageManagementGroup,
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

			// Insert AMD group after system base group (between kernel/encryption and package management)
			groups = append(groups[:1], append([]*huh.Group{amdGroup}, groups[1:]...)...)
		}
	}

	return fb.CreateForm(groups...)
}

// CreateAMDPStateForm creates AMD P-State driver selection form
// This form is only shown for AMD CPUs that support P-State
func CreateAMDPStateForm(cfg *config.Config, cpuInfo *system.CPUInfo, fb *components.FormBuilder) *huh.Form {
	// Build P-State mode options based on available modes
	var modeOptions []huh.Option[string]

	for _, mode := range cpuInfo.AMDPStateModes {
		desc := system.GetPStateModeDescription(mode)
		label := fmt.Sprintf("%s - %s", mode, desc)

		// Mark recommended mode
		if mode == cpuInfo.RecommendedPStateMode {
			label += " (recommended)"
		}

		modeOptions = append(modeOptions, huh.NewOption(label, string(mode)))
	}

	// Build description text
	zenLabel := "Unknown"
	if cpuInfo.AMDZenGen != nil {
		zenLabel = cpuInfo.AMDZenGen.Label
	}

	description := fmt.Sprintf("Detected: %s\nCPU: %s\n\nNote: Requires CPPC enabled in UEFI (AMD CBS > NBIO > SMU > CPPC)",
		zenLabel, cpuInfo.ModelName)

	return fb.CreateForm(
		huh.NewGroup(
			fb.SelectWithOptions("AMD P-State Mode", modeOptions, &cfg.AMDPState).
				Description(description),
		).Title("AMD P-State Configuration"),
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
