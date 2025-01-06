package config

import (
	"fmt"
	"os"

	"github.com/seastar-consulting/checkers/internal/errors"
	"github.com/seastar-consulting/checkers/internal/types"
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

	return &config, nil
}

// validate validates the configuration
func (m *Manager) validate(config *types.Config) error {
	if len(config.Checks) == 0 {
		return errors.NewConfigError("checks", fmt.Errorf("no checks defined"))
	}

	for _, check := range config.Checks {
		if check.Name == "" {
			return errors.NewConfigError("check.name", fmt.Errorf("check name is required"))
		}
		if check.Type == "" {
			return errors.NewConfigError("check.type", fmt.Errorf("check type is required for check %q", check.Name))
		}
	}

	return nil
}
