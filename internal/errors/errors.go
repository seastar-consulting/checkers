package errors

import "fmt"

// CheckError represents an error that occurred during check execution
type CheckError struct {
	CheckName string
	Err       error
}

func (e *CheckError) Error() string {
	return fmt.Sprintf("check %q failed: %v", e.CheckName, e.Err)
}

// NewCheckError creates a new CheckError
func NewCheckError(checkName string, err error) *CheckError {
	return &CheckError{
		CheckName: checkName,
		Err:       err,
	}
}

// ConfigError represents an error in configuration
type ConfigError struct {
	Field string
	Err   error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error in field %q: %v", e.Field, e.Err)
}

// NewConfigError creates a new ConfigError
func NewConfigError(field string, err error) *ConfigError {
	return &ConfigError{
		Field: field,
		Err:   err,
	}
}
