package types

import "encoding/json"

// Config represents the structure of the checks.yaml file
type Config struct {
	Checks []CheckItem `yaml:"checks"`
}

// CheckItem represents individual check configurations
type CheckItem struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Type        string            `yaml:"type"`
	Command     string            `yaml:"command,omitempty"`
	Parameters  map[string]string `yaml:"parameters,omitempty"`
}

// CheckResult represents the result of a single check
type CheckResult struct {
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Status string          `json:"status"`
	Output json.RawMessage `json:"output"`
	Error  string          `json:"error,omitempty"`
}
