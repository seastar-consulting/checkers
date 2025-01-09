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
func (f *Formatter) FormatResult(result types.CheckResult, checkType string, isLast bool) string {
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
	if checkType != "" {
		nameLine += fmt.Sprintf(" (%s)", checkType)
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

// contains checks if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return s == str
		}
	}
	return false
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
func (f *Formatter) FormatResults(results []types.CheckResult, checkTypes map[string]string) string {
	if len(results) == 0 {
		return ""
	}

	// Group results by type
	groups := make(map[string][]types.CheckResult)
	var groupNames []string

	for _, result := range results {
		groupName := checkTypes[result.Name]
		if groupName == "" {
			groupName = "unknown"
		}
		groups[groupName] = append(groups[groupName], result)
		if !contains(groupNames, groupName) {
			groupNames = append(groupNames, groupName)
		}
	}

	// Sort group names
	sort.Strings(groupNames)

	var output []string
	for i, groupName := range groupNames {
		isLastGroup := i == len(groupNames)-1

		// Add group header if there are multiple groups
		if len(groupNames) > 1 {
			branchSymbol := TreeBranch
			if isLastGroup {
				branchSymbol = TreeLeaf
			}
			branchPrefix := f.styles.TreeBranch.Render(branchSymbol)
			output = append(output, fmt.Sprintf("%s %s", branchPrefix, f.styles.GroupHeader.Render(groupName)))
		}

		// Add results
		groupResults := groups[groupName]
		for j, result := range groupResults {
			isLastResult := j == len(groupResults)-1 && isLastGroup
			output = append(output, f.FormatResult(result, checkTypes[result.Name], isLastResult))
		}

		// Add spacing between groups if not last
		if !isLastGroup {
			output = append(output, "")
		}
	}

	return strings.Join(output, "\n")
}
