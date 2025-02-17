package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/seastar-consulting/checkers/types"

	"github.com/seastar-consulting/checkers/internal/errors"
	"gopkg.in/yaml.v3"
)

// Manager handles configuration loading and validation
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Load loads and validates the configuration
func (m *Manager) Load() (*types.Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, errors.NewConfigError("file", err)
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.NewConfigError("parse", err)
	}

	if err := m.validate(&config); err != nil {
		return nil, err
	}

	// Expand checks with multiple items
	var expandedChecks []types.CheckItem
	for _, check := range config.Checks {
		if len(check.Items) > 0 {
			// For each item in the list, create a new check
			for i, item := range check.Items {
				// Create a copy of the check
				newCheck := types.CheckItem{
					Type:        check.Type,
					Description: check.Description,
					Command:     check.Command,
					Parameters:  item,
				}

				// If the name contains a template, render it with the item parameters
				if isTemplate(check.Name) {
					tmpl, err := template.New("check-name").Option("missingkey=error").Parse(check.Name)
					if err != nil {
						return nil, errors.NewConfigError("check.name", fmt.Errorf("invalid template in check name: %v", err))
					}

					var buf bytes.Buffer
					if err := tmpl.Execute(&buf, item); err != nil {
						return nil, errors.NewConfigError("check.name", fmt.Errorf("failed to render check name template: %v", err))
					}
					newCheck.Name = buf.String()
				} else {
					// Use the default index-based naming
					newCheck.Name = fmt.Sprintf("%s: %d", check.Name, i+1)
				}

				expandedChecks = append(expandedChecks, newCheck)
			}
		} else {
			expandedChecks = append(expandedChecks, check)
		}
	}

	config.Checks = expandedChecks
	return &config, nil
}

// validate validates the configuration
func (m *Manager) validate(config *types.Config) error {
	if len(config.Checks) == 0 {
		return errors.NewConfigError("checks", fmt.Errorf("no checks defined"))
	}

	for _, check := range config.Checks {
		// Validate required fields
		if check.Name == "" {
			return errors.NewConfigError("check.name", fmt.Errorf("check name is required"))
		}
		if check.Type == "" {
			return errors.NewConfigError("check.type", fmt.Errorf("check type is required for check %q", check.Name))
		}

		// If the name looks like a template, validate it first
		if strings.Contains(check.Name, "{{") {
			// Try to parse the template
			if _, err := template.New("check-name").Option("missingkey=error").Parse(check.Name); err != nil {
				return errors.NewConfigError("check.name", fmt.Errorf("invalid template in check name: %v", err))
			}
		}

		// Count how many of the mutually exclusive fields are set
		fieldsSet := 0
		if check.Command != "" {
			fieldsSet++
		}
		if len(check.Parameters) > 0 {
			fieldsSet++
		}
		if len(check.Items) > 0 {
			fieldsSet++
		}

		// Enforce exactly one field must be set
		if fieldsSet == 0 {
			return errors.NewConfigError("check.fields",
				fmt.Errorf("check %q must have exactly one of 'command', 'parameters', or 'items' fields", check.Name))
		}
		if fieldsSet > 1 {
			return errors.NewConfigError("check.fields",
				fmt.Errorf("check %q cannot have multiple of 'command', 'parameters', and 'items' fields", check.Name))
		}

		// If Items is used, ensure each item has parameters and validate template rendering
		if len(check.Items) > 0 {
			for i, item := range check.Items {
				if len(item) == 0 {
					return errors.NewConfigError("check.items",
						fmt.Errorf("item %d in check %q must have parameters", i, check.Name))
				}
			}

			// If the name contains a template, validate it can be rendered
			if isTemplate(check.Name) {
				tmpl, _ := template.New("check-name").Option("missingkey=error").Parse(check.Name)
				// Try to render the template with the first item to validate field access
				var buf bytes.Buffer
				if err := tmpl.Execute(&buf, check.Items[0]); err != nil {
					return errors.NewConfigError("check.name", fmt.Errorf("failed to render check name template: %v", err))
				}
			}
		}
	}

	return nil
}

// isTemplate returns true if the string contains Go template syntax
func isTemplate(s string) bool {
	return strings.Contains(s, "{{") && strings.Contains(s, "}}")
}
