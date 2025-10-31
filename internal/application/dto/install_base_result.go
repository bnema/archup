package dto

// InstallBaseResult is the result of base system installation
type InstallBaseResult struct {
	Success           bool
	PackagesInstalled []string
	ErrorDetail       string
}
