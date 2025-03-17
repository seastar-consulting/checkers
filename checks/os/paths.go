package os

import (
	"fmt"
	"os"
	"strings"

	"github.com/seastar-consulting/checkers/checks"
	"github.com/seastar-consulting/checkers/types"
)

func init() {
	checks.Register(
		"os.file_exists",
		"Check if a file exists at the given path",
		types.CheckSchema{
			Parameters: map[string]types.ParameterSchema{
				"path": {
					Type:        types.StringType,
					Description: "Path to the file to check for existence",
					Required:    true,
				},
			},
		},
		CheckFileExists,
	)

	checks.Register(
		"os.executable_exists",
		"Check if an executable exists and has proper permissions",
		types.CheckSchema{
			Parameters: map[string]types.ParameterSchema{
				"name": {
					Type:        types.StringType,
					Description: "Name of the executable to find",
					Required:    true,
				},
				"custom_path": {
					Type:        types.StringType,
					Description: "Optional custom path to look for the executable. If not specified, only PATH will be searched.",
					Required:    false,
				},
			},
		},
		CheckExecutableExists,
	)
}

// CheckFileExists checks if a file exists at the given path
func CheckFileExists(item types.CheckItem) (types.CheckResult, error) {
	path, ok := item.Parameters["path"]
	if !ok || path == "" {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  "path parameter is required",
		}, nil
	}

	_, err := os.Stat(path)
	if err == nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Success,
			Output: fmt.Sprintf("File '%s' exists", path),
		}, nil
	}
	if os.IsNotExist(err) {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Failure,
			Output: fmt.Sprintf("File '%s' does not exist", path),
		}, nil
	}
	return types.CheckResult{
		Name:   item.Name,
		Type:   item.Type,
		Status: types.Error,
		Error:  fmt.Sprintf("Error checking file '%s': %v", path, err),
	}, nil
}

// CheckExecutableExists checks if an executable exists and has proper permissions
// Parameters:
//   - name: name of the executable to find
//   - custom_path: (optional) custom path to look for the executable
func CheckExecutableExists(item types.CheckItem) (types.CheckResult, error) {
	name, ok := item.Parameters["name"]
	if !ok || name == "" {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  "name parameter is required",
		}, nil
	}

	// Check custom path first if provided
	if customPath, ok := item.Parameters["custom_path"]; ok && customPath != "" {
		fullPath := fmt.Sprintf("%s/%s", customPath, name)
		if info, err := os.Stat(fullPath); err == nil {
			if info.Mode()&0111 != 0 { // Check if executable bit is set
				return types.CheckResult{
					Name:   item.Name,
					Type:   item.Type,
					Status: types.Success,
					Output: fmt.Sprintf("Executable '%s' found at custom path '%s' with proper permissions", name, customPath),
				}, nil
			}
			return types.CheckResult{
				Name:   item.Name,
				Type:   item.Type,
				Status: types.Failure,
				Output: fmt.Sprintf("File '%s' found at custom path '%s' but lacks executable permissions", name, customPath),
			}, nil
		}
	}

	// Look in PATH
	path := os.Getenv("PATH")
	for _, dir := range strings.Split(path, ":") {
		fullPath := fmt.Sprintf("%s/%s", dir, name)
		if info, err := os.Stat(fullPath); err == nil {
			if info.Mode()&0111 != 0 { // Check if executable bit is set
				return types.CheckResult{
					Name:   item.Name,
					Type:   item.Type,
					Status: types.Success,
					Output: fmt.Sprintf("Executable '%s' found in PATH at '%s' with proper permissions", name, dir),
				}, nil
			}
		}
	}

	return types.CheckResult{
		Name:   item.Name,
		Type:   item.Type,
		Status: types.Failure,
		Output: fmt.Sprintf("Executable '%s' not found in PATH or lacks executable permissions", name),
	}, nil
}
