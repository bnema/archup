package commands

// ConfigureSystemCommand contains data for system configuration
type ConfigureSystemCommand struct {
	MountPoint  string // Root mount point where system is installed
	Hostname    string // System hostname
	Timezone    string // IANA timezone (e.g., "UTC", "Europe/Paris")
	Locale      string // System locale (e.g., "en_US.UTF-8")
	Keymap      string // Keyboard layout
	Username    string // Standard user username
	UserShell   string // Shell for user (e.g., "/bin/bash")
	UserPassword string // User password
	RootPassword string // Root password
}
