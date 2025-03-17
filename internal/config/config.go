package config

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/seastar-consulting/checkers/checks"
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

// validateParameter validates a single parameter against its schema
func (m *Manager) validateParameter(checkName string, paramName string, paramValue string, schema types.ParameterSchema) error {
	if paramValue == "" && schema.Required {
		return fmt.Errorf("parameter %q is required for check %q", paramName, checkName)
	}

	if paramValue == "" {
		return nil // Empty optional parameter is fine
	}

	switch schema.Type {
	case types.StringType:
		// All strings are valid
		return nil
	case types.BoolType:
		if _, err := strconv.ParseBool(paramValue); err != nil {
			return fmt.Errorf("parameter %q in check %q must be a boolean (true/false), got %q", paramName, checkName, paramValue)
		}
	case types.IntType:
		if _, err := strconv.ParseInt(paramValue, 10, 64); err != nil {
			return fmt.Errorf("parameter %q in check %q must be an integer, got %q", paramName, checkName, paramValue)
		}
	case types.FloatType:
		if _, err := strconv.ParseFloat(paramValue, 64); err != nil {
			return fmt.Errorf("parameter %q in check %q must be a number, got %q", paramName, checkName, paramValue)
		}
	}

	return nil
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

		// Skip parameter validation for command-type checks
		if check.Type == "command" {
			continue
		}

		// Get the check definition to validate parameters
		checkDef, err := checks.Get(check.Type)
		if err != nil {
			return errors.NewConfigError("check.type",
				fmt.Errorf("unknown check type %q for check %q", check.Type, check.Name))
		}

		// Validate parameters against schema
		if len(check.Parameters) > 0 {
			// Check for unknown parameters
			for paramName := range check.Parameters {
				if _, ok := checkDef.Schema.Parameters[paramName]; !ok {
					return errors.NewConfigError("check.parameters",
						fmt.Errorf("unknown parameter %q for check %q", paramName, check.Name))
				}
			}

			// Validate each parameter
			for paramName, schema := range checkDef.Schema.Parameters {
				paramValue := check.Parameters[paramName]
				if err := m.validateParameter(check.Name, paramName, paramValue, schema); err != nil {
					return errors.NewConfigError("check.parameters", err)
				}
			}
		} else if len(check.Items) > 0 {
			// Validate parameters in each item
			for i, item := range check.Items {
				// Check for unknown parameters
				for paramName := range item {
					if _, ok := checkDef.Schema.Parameters[paramName]; !ok {
						return errors.NewConfigError("check.items",
							fmt.Errorf("unknown parameter %q in item %d of check %q", paramName, i, check.Name))
					}
				}

				// Validate each parameter
				for paramName, schema := range checkDef.Schema.Parameters {
					paramValue := item[paramName]
					if err := m.validateParameter(check.Name, paramName, paramValue, schema); err != nil {
						return errors.NewConfigError("check.items",
							fmt.Errorf("in item %d: %v", i, err))
					}
				}
			}
		} else {
			// Check if any required parameters are missing
			for paramName, schema := range checkDef.Schema.Parameters {
				if schema.Required {
					return errors.NewConfigError("check.parameters",
						fmt.Errorf("required parameter %q is missing for check %q", paramName, check.Name))
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
