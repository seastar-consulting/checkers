package types

// CheckItem represents individual check configurations
type CheckItem struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Type        string            `yaml:"type"`
	Command     string            `yaml:"command,omitempty"`
	Parameters  map[string]string `yaml:"parameters,omitempty"`
}

// Config represents the structure of the checks.yaml file
type Config struct {
	Checks []CheckItem `yaml:"checks"`
}

// CheckResult represents the result of a single check
type CheckStatus string

const (
	Success CheckStatus = "Success"
	Failure CheckStatus = "Failure"
	Warning CheckStatus = "Warning"
	Error   CheckStatus = "Error"
)

// CheckResult represents the result of a check execution
type CheckResult struct {
	Name     string      `json:"name"`
	Status   CheckStatus `json:"status"`
	Output   string      `json:"output,omitempty"`
	Error    string      `json:"error,omitempty"`
	ExitCode int         `json:"exitCode,omitempty"`
}

// Processor defines the interface for processing check outputs
type Processor interface {
	ProcessOutput(checkName string, checkType string, output []byte) CheckResult
}
