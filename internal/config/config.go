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
	DefaultInstallDir      = "/tmp/archup-install"
	DefaultInstallPath     = ".local/share/archup/install"
	BasePackagesFile       = "base.packages"
	ExtraPackagesFile      = "extra.packages"
	LimineConfigTemplate   = "configs/limine.conf.template"
	ChaoticConfigFile      = "configs/chaotic-aur.conf"
	StarshipConfigTemplate = "configs/shell/starship.toml"
	ShellConfigTemplate  = "configs/shell/shell"
	ShellInitTemplate    = "configs/shell/init"
	ShellAliasesTemplate = "configs/shell/aliases"
	ShellEnvsTemplate    = "configs/shell/envs"
	ShellRcTemplate      = "configs/shell/rc"
	BashrcTemplate       = "configs/shell/bashrc"
)

// Encryption types
const (
	EncryptionNone    = "none"
	EncryptionLUKS    = "luks"
	EncryptionLUKSLVM = "luks-lvm"
)

// Bootloader types
const (
	BootloaderLimine     = "limine"
	BootloaderSystemdBoot = "systemd-boot"
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
	PathMnt              = "/mnt"
	PathMntEtc           = "/mnt/etc"
	PathMntEtcHostname   = "/mnt/etc/hostname"
	PathMntEtcHosts      = "/mnt/etc/hosts"
	PathMntEtcLocaleGen  = "/mnt/etc/locale.gen"
	PathMntEtcLocaleConf = "/mnt/etc/locale.conf"
	PathMntEtcVconsole   = "/mnt/etc/vconsole.conf"
	PathMntEtcSudoersD   = "/mnt/etc/sudoers.d/wheel"
	PathMntEtcSystemd    = "/mnt/etc/systemd"
	PathMntEtcSysctlD    = "/mnt/etc/sysctl.d"
	PathUsrShareZoneinfo = "/usr/share/zoneinfo"
	PathEtcLocaltime     = "/etc/localtime"
)

// Config file names
const (
	FileZramGenerator    = "zram-generator.conf"
	FileSysctlZramParams = "99-vm-zram-parameters.conf"
)

// Services
const (
	ServiceNetworkManager = "NetworkManager"
)

// User groups
const (
	GroupWheel = "wheel"
)

// Shells
const (
	ShellBash = "/bin/bash"
)

// Locale defaults
const (
	LocaleDefault     = "en_US.UTF-8"
	LocaleDefaultGen  = "en_US.UTF-8 UTF-8"
	TimezoneDefault   = ""
	KeymapDefault     = "us"
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

// Sudoers configuration
const (
	SudoersWheelContent = "%wheel ALL=(ALL:ALL) ALL\n"
	SudoersWheelPerms   = 0440
)

// Boot paths
const (
	PathMntBoot              = "/mnt/boot"
	PathMntBootEFI           = "/mnt/boot/EFI"
	PathMntBootEFILimine     = "/mnt/boot/EFI/limine"
	PathMntEtcMkinitcpio     = "/mnt/etc/mkinitcpio.conf"
	PathUsrShareLimine       = "/mnt/usr/share/limine"
	FileLimineBootloader     = "BOOTX64.EFI"
	FileLimineConfig         = "limine.conf"
)

// Mkinitcpio hooks
const (
	MkinitcpioHooksPlymouth = "HOOKS=(base udev autodetect microcode modconf kms keyboard keymap consolefont block plymouth filesystems fsck)"
	MkinitcpioHooksEncrypted = "HOOKS=(base udev autodetect microcode modconf kms keyboard keymap consolefont block plymouth encrypt filesystems fsck)"
)

// Kernel parameters
const (
	KernelParamsQuiet = "quiet splash loglevel=3 rd.udev.log_priority=3 systemd.show_status=auto"
)

// Limine configuration
const (
	LimineTimeout  = "0"
	LimineEntry    = "0"
	LimineBranding = "ArchUp"
	LimineColor    = "6"
)

// UEFI boot entry
const (
	UEFIBootLabel  = "ArchUp"
	UEFIBootLoader = "\\\\EFI\\\\limine\\\\BOOTX64.EFI"
)

// Repository paths
const (
	PathMntEtcPacmanConf = "/mnt/etc/pacman.conf"
)


// PostInstall paths
const (
	PathMntBootLogo           = "/mnt/boot/arch-logo.png"
	PathMntPlymouthThemes     = "/mnt/usr/share/plymouth/themes"
	PathMntEtcDefaultLimine   = "/mnt/etc/default/limine"
	PathMntEtcPacmanDHooks    = "/mnt/etc/pacman.d/hooks"
	PathBootEFILimineConf     = "/boot/EFI/limine/limine.conf"
	ArchLogoURL               = "assets/Arch_Linux__Crystal__icon.png"
	PlymouthThemeName         = "archup"
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
	"ufw.sh",
	"ssh-keygen.sh",
	"archup-cli.sh",
	"blesh.sh",
}

// Post-boot service template URL path
const PostBootServiceTemplate = "install/post-boot/archup-first-boot.service"

// Config holds the installation configuration
// This mirrors the shell script's config format for compatibility
type Config struct {
	// Preflight
	Hostname       string
	Username       string
	UserPassword   string
	RootPassword   string
	Email          string // Optional, for git and SSH configuration
	Keymap         string
	Timezone       string
	Locale         string
	Bootloader     string // "limine" or "systemd-boot"
	EncryptionType string // "none", "luks", or "luks-lvm"
	EncryptPassword string // Separate encryption password if different from user password

	// Form-only fields (not persisted)
	ConfirmPassword              string // Temporary field for password confirmation
	UseSamePasswordForEncryption bool   // Checkbox state for using same password for encryption

	// Partitioning
	TargetDisk     string
	BootPartition  string
	RootPartition  string
	SwapPartition  string
	EFIPartition   string
	CryptDevice    string // /dev/mapper/cryptroot if encrypted

	// Base
	KernelChoice   string // "linux", "linux-zen", "linux-lts", "linux-hardened"
	Microcode      string // "intel-ucode" or "amd-ucode"
	CPUVendor      string // "Intel", "AMD", or "Unknown"
	AMDPState      string // "active", "passive", "guided", or empty

	// Config
	NetworkManager string // "NetworkManager" by default

	// Repos
	AURHelper      string // "paru" or "yay"
	EnableMultilib bool
	EnableChaotic  bool

	// Paths
	ConfigPath     string
	LogPath        string
	InstallPath    string // ~/.local/share/archup/install
	RepoURL        string
	RawURL         string
}

// NewConfig creates a new Config with sensible defaults
// version parameter determines which branch to use for downloads (dev builds use dev branch)
func NewConfig(version string) *Config {
	// Determine which branch to use based on version
	branch := "main"
	switch {
	case version == "dev" || version == "":
		branch = "dev"
	case strings.Contains(version, "-dev"):
		branch = "dev"
	case strings.Contains(version, "-next"):
		branch = "dev"
	default:
		branch = "main"
	}

	return &Config{
		Locale:                       "en_US.UTF-8",
		Timezone:                     "UTC",
		Keymap:                       "us",
		Bootloader:                   "limine",
		EncryptionType:               "none",
		KernelChoice:                 "linux",
		NetworkManager:               "NetworkManager",
		AURHelper:                    "paru",
		EnableMultilib:               true,
		EnableChaotic:                true,
		UseSamePasswordForEncryption: true, // Default to using same password for encryption
		ConfigPath:                   DefaultConfigPath,
		LogPath:                      DefaultLogPath,
		RepoURL:                      "https://github.com/bnema/archup",
		RawURL:                       fmt.Sprintf("https://raw.githubusercontent.com/bnema/archup/%s", branch),
	}
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
	defer file.Close()

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

	return cfg, nil
}

// Save writes configuration to a shell-compatible config file
func (c *Config) Save() error {
	file, err := os.Create(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	defer file.Close()

	// Set restrictive permissions (600) since config contains passwords
	if err := os.Chmod(c.ConfigPath, 0600); err != nil {
		return fmt.Errorf("failed to set config permissions: %w", err)
	}

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header
	fmt.Fprintln(writer, "# ArchUp Installation Configuration")
	fmt.Fprintln(writer, "# This file is compatible with shell scripts")
	fmt.Fprintln(writer)

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
		{"ARCHUP_ENABLE_CHAOTIC", boolToString(c.EnableChaotic)},
	}

	for _, entry := range entries {
		if entry.value != "" {
			fmt.Fprintf(writer, "%s=\"%s\"\n", entry.key, entry.value)
		}
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
	case "ARCHUP_ENABLE_CHAOTIC":
		c.EnableChaotic = stringToBool(value)
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
