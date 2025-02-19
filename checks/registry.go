package checks

import (
	"fmt"
	"strings"
	"sync"
)

// NestedRegistry represents a nested map of checks
type NestedRegistry struct {
	Checks    map[string]*NestedRegistry
	Check     *Check
	IsPyCheck bool
}

var (
	Registry = &NestedRegistry{
		Checks: make(map[string]*NestedRegistry),
	}
	mu sync.RWMutex
)

// Register adds a new check to the registry
func Register(name, description string, fn CheckFunc) {
	mu.Lock()
	defer mu.Unlock()

	parts := strings.Split(name, ".")
	current := Registry

	// Special case for py namespace
	if parts[0] == "py" {
		if _, exists := current.Checks["py"]; !exists {
			current.Checks["py"] = &NestedRegistry{
				IsPyCheck: true,
				Check: &Check{
					Name:        "py",
					Description: "Python function checks",
					Func:        fn,
				},
			}
		}
		return
	}

	// For all other namespaces, create nested structure
	for i, part := range parts {
		if _, exists := current.Checks[part]; !exists {
			current.Checks[part] = &NestedRegistry{
				Checks: make(map[string]*NestedRegistry),
			}
		}
		current = current.Checks[part]

		// Set the check at the leaf node
		if i == len(parts)-1 {
			current.Check = &Check{
				Name:        name,
				Description: description,
				Func:        fn,
			}
		}
	}
}

// Get returns a registered check
func Get(name string) (Check, error) {
	mu.RLock()
	defer mu.RUnlock()

	parts := strings.Split(name, ".")
	current := Registry

	// Special case for py namespace
	if parts[0] == "py" {
		if pyCheck, exists := current.Checks["py"]; exists && pyCheck.IsPyCheck {
			check := *pyCheck.Check
			check.Name = name // Use the full name for the check
			return check, nil
		}
		return Check{}, fmt.Errorf("python check handler not found")
	}

	// For other namespaces, traverse the tree
	for _, part := range parts {
		if next, exists := current.Checks[part]; exists {
			current = next
		} else {
			return Check{}, fmt.Errorf("check %s not found", name)
		}
	}

	if current.Check == nil {
		return Check{}, fmt.Errorf("check %s not found", name)
	}

	return *current.Check, nil
}

// List returns all registered checks
func List() []Check {
	mu.RLock()
	defer mu.RUnlock()
	checks := make([]Check, 0)
	listChecksRecursive(Registry, &checks)
	return checks
}

// listChecksRecursive recursively traverses the registry tree to collect all checks
func listChecksRecursive(reg *NestedRegistry, checks *[]Check) {
	if reg.Check != nil {
		*checks = append(*checks, *reg.Check)
	}
	for _, child := range reg.Checks {
		listChecksRecursive(child, checks)
	}
}
