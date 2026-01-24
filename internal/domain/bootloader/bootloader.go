package bootloader

import "errors"

// BootloaderType represents the bootloader implementation
type BootloaderType int

const (
	// BootloaderTypeLimine is the Limine bootloader
	BootloaderTypeLimine BootloaderType = iota

	// BootloaderTypeSystemdBoot is systemd-boot
	BootloaderTypeSystemdBoot
)

// String returns human-readable bootloader name
func (b BootloaderType) String() string {
	switch b {
	case BootloaderTypeSystemdBoot:
		return "systemd-boot"
	default:
		return "Limine"
	}
}

// Bootloader is an immutable value object representing bootloader configuration
type Bootloader struct {
	bootType BootloaderType
	timeout  int // seconds
	branding string
	efiPath  string
}

// NewBootloader creates a new Bootloader value object with validation
func NewBootloader(bootType BootloaderType, timeoutSeconds int, branding string) (*Bootloader, error) {
	// Validate timeout
	if timeoutSeconds < 0 {
		return nil, errors.New("bootloader timeout cannot be negative")
	}

	if timeoutSeconds > 600 {
		return nil, errors.New("bootloader timeout cannot exceed 600 seconds")
	}

	// Validate branding
	if branding == "" {
		return nil, errors.New("bootloader branding cannot be empty")
	}

	if len(branding) > 64 {
		return nil, errors.New("bootloader branding too long")
	}

	return &Bootloader{
		bootType: bootType,
		timeout:  timeoutSeconds,
		branding: branding,
		efiPath:  "/boot/efi",
	}, nil
}

// Type returns the bootloader type
func (b *Bootloader) Type() BootloaderType {
	return b.bootType
}

// Timeout returns the bootloader timeout in seconds
func (b *Bootloader) Timeout() int {
	return b.timeout
}

// Branding returns the bootloader branding string
func (b *Bootloader) Branding() string {
	return b.branding
}

// EFIPath returns the EFI boot path
func (b *Bootloader) EFIPath() string {
	return b.efiPath
}

// IsLimine returns true if bootloader is Limine
func (b *Bootloader) IsLimine() bool {
	return b.bootType == BootloaderTypeLimine
}

// IsSystemdBoot returns true if bootloader is systemd-boot
func (b *Bootloader) IsSystemdBoot() bool {
	return b.bootType == BootloaderTypeSystemdBoot
}

// String returns human-readable representation
func (b *Bootloader) String() string {
	return "Bootloader(type=" + b.bootType.String() + ", timeout=" + intToString(b.timeout) + "s)"
}

// Equals checks if two Bootloader objects are equal
func (b *Bootloader) Equals(other *Bootloader) bool {
	if other == nil {
		return false
	}
	return b.bootType == other.bootType &&
		b.timeout == other.timeout &&
		b.branding == other.branding
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToString(-n)
	}
	digits := "0123456789"
	var result string
	for n > 0 {
		result = string(digits[n%10]) + result
		n /= 10
	}
	return result
}
