package types

import "time"

// CheckItem represents a single check to be executed
type CheckItem struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description,omitempty"`
	Type        string              `yaml:"type"`
	Command     string              `yaml:"command,omitempty"`
	Parameters  map[string]string   `yaml:"parameters,omitempty"`
	Items       []map[string]string `yaml:"items,omitempty"`
}

// Config represents the structure of the checks.yaml file
type Config struct {
	Timeout *time.Duration `yaml:"timeout,omitempty"`
	Checks  []CheckItem    `yaml:"checks"`
}

// CheckStatus represents the result of a single check
type CheckStatus string

const (
	Success CheckStatus = "Success"
	Failure CheckStatus = "Failure"
	Warning CheckStatus = "Warning"
	Error   CheckStatus = "Error"
)

type CheckResult struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Status CheckStatus `json:"status"`
	Output string      `json:"output"`
	Error  string      `json:"error,omitempty"`
}
