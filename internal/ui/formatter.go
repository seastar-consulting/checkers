package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/seastar-consulting/checkers/internal/types"
)

// Formatter handles the formatting of check results
type Formatter struct {
	styles  *Styles
	verbose bool
}

// NewFormatter creates a new Formatter instance
func NewFormatter(verbose bool) *Formatter {
	return &Formatter{
		styles:  NewStyles(),
		verbose: verbose,
	}
}

// FormatResult formats a single check result
func (f *Formatter) FormatResult(result types.CheckResult) string {
	var icon string
	var nameStyle lipgloss.Style

	switch result.Status {
	case types.Success:
		icon = CheckPassIcon
		nameStyle = f.styles.Success
	case types.Failure:
		icon = CheckFailIcon
		nameStyle = f.styles.Error
	default:
		icon = CheckErrorIcon
		nameStyle = f.styles.Warning
	}

	// Format the name line
	nameLine := fmt.Sprintf("%s %s", icon, nameStyle.Render(result.Name))
	if result.Type != "" {
		nameLine += fmt.Sprintf(" (%s)", result.Type)
	}

	var output []string
	output = append(output, nameLine)

	// Add output box if verbose mode is on
	if result.Output != "" && f.verbose {
		output = append(output, f.styles.OutputBox.Render(fmt.Sprintf("Output: %s", result.Output)))
	}

	// Add error box - first line always shown in red, rest in grey if verbose
	if result.Error != "" {
		lines := strings.Split(strings.TrimSpace(result.Error), "\n")
		if len(lines) > 0 {
			// First line always in error box
			output = append(output, f.styles.ErrorBox.Render(fmt.Sprintf("Error: %s", lines[0])))

			// Rest of the lines in grey box when verbose
			if len(lines) > 1 && f.verbose {
				restOfError := strings.Join(lines[1:], "\n")
				output = append(output, f.styles.OutputBox.Render(restOfError))
			}
		}
	}

	return strings.Join(output, "\n")
}

// FormatResults formats multiple check results
func (f *Formatter) FormatResults(results []types.CheckResult) string {
	var output []string
	for _, result := range results {
		output = append(output, f.FormatResult(result))
	}
	return strings.Join(output, "\n") + "\n"
}
