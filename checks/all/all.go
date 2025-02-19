// Package all imports all available checks to ensure they are registered
package all

import (
	_ "github.com/seastar-consulting/checkers/checks/cloud" // Register cloud checks
	_ "github.com/seastar-consulting/checkers/checks/k8s"   // Register k8s checks
	_ "github.com/seastar-consulting/checkers/checks/os"    // Register os checks
	_ "github.com/seastar-consulting/checkers/checks/py"    // Register python checks
	// Add new check packages here
)
