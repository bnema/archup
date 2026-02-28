package dto

// RepositoriesResult is the result of repository configuration
type RepositoriesResult struct {
	Success     bool
	Multilib    bool
	AURHelper   string
	ErrorDetail string
}
