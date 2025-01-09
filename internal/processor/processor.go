package processor

import (
	"encoding/json"
	"strings"

	"github.com/seastar-consulting/checkers/internal/types"
)

// Processor handles the processing of check outputs
type Processor struct{}

// NewProcessor creates a new Processor instance
func NewProcessor() *Processor {
	return &Processor{}
}

// ProcessOutput processes the raw output from a check execution
func (p *Processor) ProcessOutput(checkName string, checkType string, rawOutput []byte) types.CheckResult {
	result := types.CheckResult{
		Name: checkName,
	}

	// Try to unmarshal the output as JSON
	var output map[string]interface{}
	if err := json.Unmarshal(rawOutput, &output); err != nil {
		result.Status = types.Error
		result.Error = err.Error()
		return result
	}

	// Check for error first
	if errStr, ok := output["error"].(string); ok && errStr != "" {
		result.Status = types.Error
		result.Error = errStr
		return result
	}

	// Process status
	if status, ok := output["status"].(string); ok {
		switch strings.ToLower(status) {
		case "success", "pass":
			result.Status = types.Success
		case "failure", "fail":
			result.Status = types.Failure
		case "warning", "warn":
			result.Status = types.Warning
		default:
			result.Status = types.Error
		}
	} else {
		result.Status = types.Error
		result.Error = "missing status field"
		return result
	}

	// Process output
	if outputStr, ok := output["output"].(string); ok {
		result.Output = outputStr
	}

	return result
}
