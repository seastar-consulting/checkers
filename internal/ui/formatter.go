package ui

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/seastar-consulting/checkers/types"

	"github.com/charmbracelet/lipgloss"
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

// formatResult formats a single check result
func (f *Formatter) formatResult(result types.CheckResult, isLast bool) string {
	var icon string
	var nameStyle lipgloss.Style

	switch result.Status {
	case types.Success:
		icon = CheckPassIcon
		nameStyle = f.styles.Success
	case types.Failure:
		icon = CheckFailIcon
		nameStyle = f.styles.Error
	case types.Error:
		icon = CheckErrorIcon
		nameStyle = f.styles.Error
	case types.Warning:
		icon = CheckWarningIcon
		nameStyle = f.styles.Warning
	default:
		icon = CheckErrorIcon
		nameStyle = f.styles.Error
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

// FormatFunc defines the interface for result formatting functions
type FormatFunc func([]types.CheckResult, types.OutputMetadata) string

// FormatResultsPretty formats multiple check results in a pretty format
func (f *Formatter) FormatResultsPretty(results []types.CheckResult, metadata types.OutputMetadata) string {
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
			output = append(output, f.formatResult(result, isLastResult))
		}

		// Add spacing between groups if not last
		if !isLastGroup {
			output = append(output, "")
		}
	}

	return strings.Join(output, "\n") + "\n\n"
}

// FormatResultsJSON formats check results as JSON
func (f *Formatter) FormatResultsJSON(results []types.CheckResult, metadata types.OutputMetadata) string {
	output := types.JSONOutput{
		Results:  results,
		Metadata: metadata,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal results: %v"}`, err)
	}

	return string(jsonBytes)
}

// HTMLData represents the data passed to the HTML template
type HTMLData struct {
	Groups   map[string][]types.CheckResult
	Metadata types.OutputMetadata
}

// FormatResultsHTML formats check results as HTML
func (f *Formatter) FormatResultsHTML(results []types.CheckResult, metadata types.OutputMetadata) string {
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

	// Sort results within each group by name
	for groupName, groupResults := range groups {
		sort.Slice(groupResults, func(i, j int) bool {
			return groupResults[i].Name < groupResults[j].Name
		})
		groups[groupName] = groupResults
	}

	// Prepare data for template
	data := HTMLData{
		Groups:   groups,
		Metadata: metadata,
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"toLowerString": func(v interface{}) string {
			return strings.ToLower(fmt.Sprintf("%v", v))
		},
	}

	// Get the path to the template file
	_, currentFilePath, _, _ := runtime.Caller(0)
	templateDir := filepath.Dir(currentFilePath)
	templatePath := filepath.Join(templateDir, "templates", "results.html.tmpl")

	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Fall back to embedded template if file doesn't exist
		return fmt.Sprintf("<html><body><h1>Error</h1><p>Template file not found: %s</p></body></html>", templatePath)
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		return fmt.Sprintf("<html><body><h1>Error</h1><p>Failed to parse HTML template: %v</p></body></html>", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("<html><body><h1>Error</h1><p>Failed to execute HTML template: %v</p></body></html>", err)
	}

	return buf.String()
}
