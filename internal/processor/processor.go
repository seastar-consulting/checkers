package processor

import (
	"fmt"
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
func (p *Processor) ProcessOutput(checkName string, checkType string, output map[string]interface{}) types.CheckResult {
	result := types.CheckResult{
		Name: checkName,
		Type: checkType,
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
			result.Error = fmt.Sprintf("unknown status: %s", status)
		}
	} else if output["output"] != nil {
		// If there's output but no status, consider it a success
		result.Status = types.Success
	} else {
		result.Status = types.Error
		result.Error = "no status or output provided"
	}

	// Process output
	if output, ok := output["output"].(string); ok {
		result.Output = output
	}

	return result
}
