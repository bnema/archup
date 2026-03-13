package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	DefaultConfigPath = "/var/log/archup-install.conf"
	DefaultLogPath    = "/var/log/archup-install.log"

	// Install paths
	DefaultInstallDir     = "/tmp/archup-install"
	DefaultInstallRepoDir = "/tmp/archup-install/repo"
	DefaultInstallPath    = ".local/share/archup/install"
	BasePackagesFile      = "base.packages"
	ExtraPackagesFile     = "extra.packages"
	LimineConfigTemplate  = "configs/limine.conf.template"
)

// Encryption types
const (
	EncryptionNone    = "none"
	EncryptionLUKS    = "luks"
	EncryptionLUKSLVM = "luks-lvm"
)

// Bootloader types
const (
	BootloaderLimine = "limine"
)

// Kernel choices
const (
	KernelLinux         = "linux"
	KernelLinuxZen      = "linux-zen"
	KernelLinuxLTS      = "linux-lts"
	KernelLinuxHardened = "linux-hardened"
	KernelLinuxCachyOS  = "linux-cachyos"
)

// System paths
const (
	PathMnt = "/mnt"
)

// Mkinitcpio hooks
const (
	MkinitcpioHooksPlymouth  = "HOOKS=(base udev autodetect microcode modconf kms keyboard keymap consolefont block plymouth filesystems fsck)"
	MkinitcpioHooksEncrypted = "HOOKS=(base udev autodetect microcode modconf kms keyboard keymap consolefont block plymouth encrypt filesystems fsck)"
)

// Kernel parameters
const (
	KernelParamsQuiet = "quiet splash loglevel=3 rd.udev.log_priority=3 systemd.show_status=auto"
)

// Limine configuration
const (
	LimineColor = "6"
)

// UEFI boot entry
const (
	UEFIBootLabel  = "Arch Linux"
	UEFIBootLoader = "\\EFI\\limine\\BOOTX64.EFI"
)

// PostInstall paths
const (
	PathMntPlymouthThemes  = "/mnt/usr/share/plymouth/themes"
	PathMntEtcPacmanDHooks = "/mnt/etc/pacman.d/hooks"
	PathBootLimineConf     = "/boot/limine.conf"
	PlymouthThemeName      = "archup"
)

// Zram configuration
const (
	ZramConfigContent = `[zram0]
zram-size = min(ram / 2, 4096)
compression-algorithm = zstd
`
	ZramSysctlContent = `vm.swappiness = 180
vm.watermark_boost_factor = 0
vm.watermark_scale_factor = 125
vm.page-cluster = 0
`
)

// Zram file names
const (
	FileZramGenerator    = "zram-generator.conf"
	FileSysctlZramParams = "99-vm-zram-parameters.conf"
)

// Sudoers configuration
const (
	SudoersWheelContent = "%wheel ALL=(ALL:ALL) ALL\n"
	SudoersWheelPerms   = 0440
)

// Plymouth files
var PlymouthFiles = []string{
	"archup.plymouth",
	"archup.script",
	"logo.png",
	"bullet.png",
	"entry.png",
	"lock.png",
	"progress_bar.png",
	"progress_box.png",
}

// Post-boot paths and files
const (
	PathMntPostBoot      = "/mnt/usr/local/share/archup/post-boot"
	PathMntSystemdSystem = "/mnt/etc/systemd/system"
	PostBootServiceName  = "archup-first-boot.service"
)

// Post-boot script files to download
var PostBootScripts = []string{
	"all.sh",
	"snapper.sh",
	"firewalld.sh",
	"ssh-keygen.sh",
	"archup-cli.sh",
	"blesh.sh",
	"cli-tools.sh",
	"dms-opt-in.sh",
}

// Post-boot service template URL path
const PostBootServiceTemplate = "install/mandatory/post-boot/archup-first-boot.service"

// Config holds the installation configuration
// This mirrors the shell script's config format for compatibility
type Config struct {
	// Preflight
	Hostname        string
	Username        string
	UserPassword    string
	RootPassword    string
	Email           string // Optional, for git and SSH configuration
	Keymap          string
	Timezone        string
	Locale          string
	Bootloader      string // "limine" (only supported bootloader)
	EncryptionType  string // "none", "luks", or "luks-lvm"
	EncryptPassword string // Separate encryption password if different from user password

	// Form-only fields (not persisted)
	ConfirmPassword              string // Temporary field for password confirmation
	UseSamePasswordForEncryption bool   // Checkbox state for using same password for encryption

	// Partitioning
	TargetDisk    string
	BootPartition string
	RootPartition string
	SwapPartition string
	EFIPartition  string
	CryptDevice   string // /dev/mapper/cryptroot if encrypted

	// Base
	KernelChoice string // "linux", "linux-zen", "linux-lts", "linux-hardened"
	Microcode    string // "intel-ucode" or "amd-ucode"
	CPUVendor    string // "Intel", "AMD", or "Unknown"
	AMDPState    string // "active", "passive", "guided", or empty

	// Config
	NetworkManager string // "NetworkManager" by default

	// Repos
	AURHelper      string // "paru" or "yay"
	EnableMultilib bool

	// Paths
	ConfigPath  string
	LogPath     string
	InstallPath string // ~/.local/share/archup/install
	RepoURL     string
	RawURL      string
}

// NewConfig creates a new Config with sensible defaults.
// version determines which git ref to use for downloads and bootstrap assets.
// ENV=dev forces the dev ref regardless of the version string.
func NewConfig(version string) *Config {
	ref := bootstrapRef(version)

	return &Config{
		Hostname:                     "arch",
		Locale:                       "en_US.UTF-8",
		Timezone:                     "UTC",
		Keymap:                       "us",
		Bootloader:                   "limine",
		EncryptionType:               "none",
		KernelChoice:                 "linux",
		NetworkManager:               "NetworkManager",
		AURHelper:                    "paru",
		EnableMultilib:               true,
		UseSamePasswordForEncryption: true, // Default to using same password for encryption
		ConfigPath:                   DefaultConfigPath,
		LogPath:                      DefaultLogPath,
		RepoURL:                      "https://github.com/bnema/archup",
		RawURL:                       fmt.Sprintf("https://raw.githubusercontent.com/bnema/archup/%s", ref),
	}
}

func bootstrapRef(version string) string {
	if os.Getenv("ENV") == "dev" {
		return "dev"
	}

	version = strings.TrimSpace(version)
	switch {
	case version == "" || version == "dev":
		return "dev"
	case strings.Contains(version, "-dev"):
		return "dev"
	case strings.Contains(version, "-next"):
		return "dev"
	default:
		return version
	}
}

// Branch returns the git ref to use for asset downloads and repo cloning.
func (c *Config) Branch() string {
	rawURL := strings.TrimSpace(c.RawURL)
	if rawURL == "" {
		return "main"
	}
	// Extract branch from RawURL: "https://raw.githubusercontent.com/bnema/archup/<branch>"
	parts := strings.Split(rawURL, "/")
	last := parts[len(parts)-1]
	if last == "" {
		return "main"
	}
	return last
}

// Load reads configuration from a shell-compatible config file
func Load(path string, version string) (*Config, error) {
	cfg := NewConfig(version)
	cfg.ConfigPath = path

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return defaults if file doesn't exist
		}
		return nil, fmt.Errorf("failed to open config: %w", err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`) // Remove quotes

		cfg.setValue(key, value)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close config file: %w", err)
	}

	return cfg, nil
}

// Save writes configuration to a shell-compatible config file
func (c *Config) Save() error {
	file, err := os.Create(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	// Set restrictive permissions (600) since config contains passwords
	if err := os.Chmod(c.ConfigPath, 0600); err != nil {
		return fmt.Errorf("failed to set config permissions: %w", err)
	}

	writer := bufio.NewWriter(file)

	// Write header
	if _, err := fmt.Fprintln(writer, "# ArchUp Installation Configuration"); err != nil {
		return fmt.Errorf("failed to write config header: %w", err)
	}
	if _, err := fmt.Fprintln(writer, "# This file is compatible with shell scripts"); err != nil {
		return fmt.Errorf("failed to write config header: %w", err)
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return fmt.Errorf("failed to write config header: %w", err)
	}

	// Write all values
	entries := []struct {
		key   string
		value string
	}{
		{"ARCHUP_HOSTNAME", c.Hostname},
		{"ARCHUP_USERNAME", c.Username},
		{"ARCHUP_USER_PASSWORD", c.UserPassword},
		{"ARCHUP_ROOT_PASSWORD", c.RootPassword},
		{"ARCHUP_KEYMAP", c.Keymap},
		{"ARCHUP_TIMEZONE", c.Timezone},
		{"ARCHUP_LOCALE", c.Locale},
		{"ARCHUP_BOOTLOADER", c.Bootloader},
		{"ARCHUP_ENCRYPTION", c.EncryptionType},
		{"ARCHUP_ENCRYPTION_PASSWORD", c.EncryptPassword},
		{"ARCHUP_TARGET_DISK", c.TargetDisk},
		{"ARCHUP_BOOT_PARTITION", c.BootPartition},
		{"ARCHUP_ROOT_PARTITION", c.RootPartition},
		{"ARCHUP_SWAP_PARTITION", c.SwapPartition},
		{"ARCHUP_EFI_PARTITION", c.EFIPartition},
		{"ARCHUP_KERNEL", c.KernelChoice},
		{"ARCHUP_AMD_PSTATE", c.AMDPState},
		{"ARCHUP_NETWORK_MANAGER", c.NetworkManager},
		{"ARCHUP_AUR_HELPER", c.AURHelper},
		{"ARCHUP_ENABLE_MULTILIB", boolToString(c.EnableMultilib)},
	}

	for _, entry := range entries {
		if entry.value != "" {
			if _, err := fmt.Fprintf(writer, "%s=\"%s\"\n", entry.key, entry.value); err != nil {
				return fmt.Errorf("failed to write config entry: %w", err)
			}
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush config writer: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close config file: %w", err)
	}

	return nil
}

// setValue sets a config value based on key name
func (c *Config) setValue(key, value string) {
	switch key {
	case "ARCHUP_HOSTNAME":
		c.Hostname = value
	case "ARCHUP_USERNAME":
		c.Username = value
	case "ARCHUP_USER_PASSWORD":
		c.UserPassword = value
	case "ARCHUP_ROOT_PASSWORD":
		c.RootPassword = value
	case "ARCHUP_KEYMAP":
		c.Keymap = value
	case "ARCHUP_TIMEZONE":
		c.Timezone = value
	case "ARCHUP_LOCALE":
		c.Locale = value
	case "ARCHUP_BOOTLOADER":
		c.Bootloader = value
	case "ARCHUP_ENCRYPTION":
		c.EncryptionType = value
	case "ARCHUP_ENCRYPTION_PASSWORD":
		c.EncryptPassword = value
	case "ARCHUP_TARGET_DISK":
		c.TargetDisk = value
	case "ARCHUP_BOOT_PARTITION":
		c.BootPartition = value
	case "ARCHUP_ROOT_PARTITION":
		c.RootPartition = value
	case "ARCHUP_SWAP_PARTITION":
		c.SwapPartition = value
	case "ARCHUP_EFI_PARTITION":
		c.EFIPartition = value
	case "ARCHUP_KERNEL":
		c.KernelChoice = value
	case "ARCHUP_AMD_PSTATE":
		c.AMDPState = value
	case "ARCHUP_NETWORK_MANAGER":
		c.NetworkManager = value
	case "ARCHUP_AUR_HELPER":
		c.AURHelper = value
	case "ARCHUP_ENABLE_MULTILIB":
		c.EnableMultilib = stringToBool(value)
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func stringToBool(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}
