package dto

// BootstrapResult contains the result of the bootstrap phase
type BootstrapResult struct {
	Success     bool
	InstallDir  string
	Method      string // "clone" or "download"
	ErrorDetail string
}
