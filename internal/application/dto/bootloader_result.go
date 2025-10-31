package dto

// BootloaderResult is the result of bootloader installation
type BootloaderResult struct {
	Success        bool
	BootloaderType string
	Timeout        int
	ErrorDetail    string
}
