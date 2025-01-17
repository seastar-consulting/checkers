package os

import (
	"fmt"
	"os"

	"github.com/seastar-consulting/checkers/checks"
	"github.com/seastar-consulting/checkers/types"
)

func init() {
	checks.Register("os.file_exists", "Check if a file exists at the given path", CheckFileExists)
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
