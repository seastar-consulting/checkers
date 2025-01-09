package ui

import (
	"fmt"
	"sort"
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
func (f *Formatter) FormatResult(result types.CheckResult, isLast bool) string {
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

	// Format the name line with tree branch
	branchSymbol := TreeBranch
	if isLast {
		branchSymbol = TreeLeaf
	}
	branchPrefix := f.styles.TreeBranch.Render(branchSymbol)
	nameLine := fmt.Sprintf("%s %s %s", branchPrefix, icon, nameStyle.Render(result.Name))
	if result.Type != "" {
		nameLine += fmt.Sprintf(" (%s)", result.Type)
	}

	var output []string
	output = append(output, nameLine)

	// Add output box if verbose mode is on
	if result.Output != "" && f.verbose {
		if isLast {
			output = append(output, f.styles.OutputBox.Render(result.Output))
		} else {
			verticalBar := f.styles.TreeBranch.Render(TreeVertical)
			output = append(output, prepend(f.styles.OutputBox.Render(result.Output), verticalBar)...)
		}
	}

	// Add error box - first line always shown in red, rest in grey if verbose
	if result.Error != "" {
		lines := strings.Split(strings.TrimSpace(result.Error), "\n")
		if len(lines) > 0 {
			verticalBar := f.styles.TreeBranch.Render(TreeVertical)
			// First line always in error box
			if isLast {
				output = append(output, f.styles.ErrorBox.Render(lines[0]))
			} else {
				output = append(output, prepend(f.styles.ErrorBox.Render(lines[0]), verticalBar)...)
			}

			// Rest of the lines in grey box when verbose
			if len(lines) > 1 && f.verbose {
				restOfError := strings.Join(lines[1:], "\n")
				if isLast {
					output = append(output, f.styles.OutputBox.Render(restOfError))
				} else {
					output = append(output, prepend(f.styles.OutputBox.Render(restOfError), verticalBar)...)
				}
			}
		}
	}

	return strings.Join(output, "\n")
}

// prepend adds a prefix to each line of a string
func prepend(box string, item string) []string {
	lines := strings.Split(box, "\n")
	for j := 0; j < len(lines); j++ {
		if len(lines[j]) > 0 && lines[j][0] == ' ' {
			lines[j] = item + lines[j][1:]
		} else {
			lines[j] = item + lines[j]
		}
	}
	return lines
}

// FormatResults formats multiple check results
func (f *Formatter) FormatResults(results []types.CheckResult) string {
	// Group results by type
	groups := make(map[string][]types.CheckResult)

	for _, result := range results {
		groupKey := "command"
		if result.Type != "command" {
			// For native checks, use the top-level package as the group
			parts := strings.Split(result.Type, ".")
			if len(parts) > 0 {
				groupKey = parts[0]
			}
		}
		groups[groupKey] = append(groups[groupKey], result)
	}

	// Get sorted group names for consistent output
	var groupNames []string
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	var output []string
	isLastGroup := false
	for i, groupName := range groupNames {
		isLastGroup = i == len(groupNames)-1

		// Add group header
		output = append(output, f.styles.GroupHeader.Render(strings.ToUpper(groupName)))

		// Add results for this group
		groupResults := groups[groupName]
		for j, result := range groupResults {
			isLastResult := j == len(groupResults)-1
			output = append(output, f.FormatResult(result, isLastResult))
		}

		// Add spacing between groups if not last
		if !isLastGroup {
			output = append(output, "")
		}
	}

	return strings.Join(output, "\n") + "\n\n"
}
