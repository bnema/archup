package dto

// ConfigureSystemResult is the result of system configuration
type ConfigureSystemResult struct {
	Success     bool
	Hostname    string
	Timezone    string
	Username    string
	ErrorDetail string
}
