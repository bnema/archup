package dto

// PostInstallResult is the result of post-installation tasks
type PostInstallResult struct {
	Success              bool
	TasksRun             []string
	ErrorDetail          string
	VerificationWarnings []string
}
