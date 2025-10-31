package dto

// RepositoriesResult is the result of repository configuration
type RepositoriesResult struct {
	Success     bool
	Multilib    bool
	Chaotic     bool
	AURHelper   string
	ErrorDetail string
}
