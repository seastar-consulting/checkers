package checks

import (
	"fmt"
	"sync"
)

var (
	Registry = make(map[string]Check)
	mu       sync.RWMutex
)

// Register adds a new check to the registry
func Register(name, description string, fn CheckFunc) {
	mu.Lock()
	defer mu.Unlock()
	Registry[name] = Check{
		Name:        name,
		Description: description,
		Func:        fn,
	}
}

// Get returns a registered check
func Get(name string) (Check, error) {
	mu.RLock()
	defer mu.RUnlock()
	check, ok := Registry[name]
	if !ok {
		return Check{}, fmt.Errorf("check %s not found", name)
	}
	return check, nil
}

// List returns all registered checks
func List() []Check {
	mu.RLock()
	defer mu.RUnlock()
	checks := make([]Check, 0, len(Registry))
	for _, check := range Registry {
		checks = append(checks, check)
	}
	return checks
}
