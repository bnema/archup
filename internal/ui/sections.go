package ui

// Form section descriptions (DRY: centralized text)
const (
	// Welcome
	WelcomeTagline = "A minimal, slightly opinionated Arch Linux installer."

	// Preflight sections
	SectionUserIdentity = "Configure your user account and system details"
	DescEmail           = "Email is used for SSH key generation and git configuration"
	DescTimezone        = "e.g. America/New_York, Europe/London, Asia/Tokyo"
	DescLocale          = "e.g. en_US.UTF-8, en_GB.UTF-8, de_DE.UTF-8"
	DescKeymap          = "e.g. us, uk, de, fr"

	// Disk selection
	SectionDiskSelection = "Select installation disk (ALL DATA WILL BE ERASED)"
	WarnDiskErase        = "WARNING: This will delete all data on the selected disk!"

	// Encryption
	SectionDiskEncryption   = "Disk Encryption"
	DescEncryption          = "LUKS encryption protects your data with Argon2id (2000ms iteration)"
	DescEncryptionPassword  = "Leave empty to use the same password as your user account"
	DescEncryptionSamePass  = "Use the same password for disk encryption as your user account?"

	// Kernel
	SectionKernel     = "Kernel Selection"
	DescKernelLinux   = "Stable mainline kernel (recommended)"
	DescKernelLTS     = "Long-term support (maximum stability)"
	DescKernelZen     = "Performance-optimized for general use"
	DescKernelHardened = "Security-focused kernel"
	DescKernelCachyOS = "Gaming-optimized (requires CachyOS repo)"

	// AMD P-State
	SectionAMDTuning = "AMD CPU Tuning"
	DescAMDPState    = "Select AMD P-State driver mode for CPU frequency scaling"

	// Repository options
	DescMultilib     = "Required for 32-bit applications and Wine"
	DescChaoticAUR   = "Community repository with pre-built AUR packages"
	DescAURHelper    = "Tool for installing packages from Arch User Repository"
	DescAURHelperParu = "Faster with more features (recommended)"
	DescAURHelperYay  = "Classic, widely compatible"
)
