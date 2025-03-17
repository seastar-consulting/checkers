package checks

import (
	"fmt"
	"sync"

	"github.com/seastar-consulting/checkers/types"
)

var (
	Registry = make(map[string]types.CheckDefinition)
	mu       sync.RWMutex
)

// Register adds a new check to the registry
func Register(name, description string, schema types.CheckSchema, fn types.CheckFunc) {
	mu.Lock()
	defer mu.Unlock()
	Registry[name] = types.CheckDefinition{
		Name:        name,
		Description: description,
		Schema:      schema,
		Handler:     fn,
	}
}

// Get returns a registered check
func Get(name string) (types.CheckDefinition, error) {
	mu.RLock()
	defer mu.RUnlock()
	check, ok := Registry[name]
	if !ok {
		return types.CheckDefinition{}, fmt.Errorf("check %s not found", name)
	}
	return check, nil
}

// List returns all registered checks
func List() []types.CheckDefinition {
	mu.RLock()
	defer mu.RUnlock()
	checks := make([]types.CheckDefinition, 0, len(Registry))
	for _, check := range Registry {
		checks = append(checks, check)
	}
	return checks
}
