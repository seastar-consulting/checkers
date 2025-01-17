package os

import (
	"fmt"
	"os"

	"github.com/seastar-consulting/checkers/checks"
	"github.com/seastar-consulting/checkers/types"
)

func init() {
	checks.Register("os.file_exists", "Check if a file exists", FileExists)
}

// FileExists checks if a file exists at the given path
func FileExists(params map[string]interface{}) (types.CheckResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return types.CheckResult{
			Status: types.Error,
			Error:  "path parameter is required",
			Output: "",
		}, fmt.Errorf("path parameter is required")
	}

	_, err := os.Stat(path)
	if err == nil {
		return types.CheckResult{
			Name:   "os.file_exists",
			Type:   "File Check",
			Status: types.Success,
			Output: fmt.Sprintf("File %s exists", path),
		}, nil
	}
	if os.IsNotExist(err) {
		return types.CheckResult{
			Name:   "os.file_exists",
			Type:   "File Check",
			Status: types.Failure,
			Output: fmt.Sprintf("File %s does not exist", path),
		}, nil
	}
	return types.CheckResult{
		Name:   "os.file_exists",
		Type:   "File Check",
		Status: types.Error,
		Output: "",
		Error:  err.Error(),
	}, err
}
