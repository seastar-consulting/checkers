// Package all imports all available checks to ensure they are registered
package all

import (
	_ "github.com/seastar-consulting/checkers/checks/cloud" // Register cloud checks
	_ "github.com/seastar-consulting/checkers/checks/git"   // Register git checks
	_ "github.com/seastar-consulting/checkers/checks/k8s"   // Register k8s checks
	_ "github.com/seastar-consulting/checkers/checks/os"    // Register os checks
	// Add new check packages here
)
