package os

import (
	"fmt"
	"os"

	"github.com/seastar-consulting/checkers/checks"
)

func init() {
	checks.Register("os.file_exists", "Check if a file exists", FileExists)
}

// FileExists checks if a file exists at the given path
func FileExists(params map[string]interface{}) (map[string]interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	_, err := os.Stat(path)
	if err == nil {
		return map[string]interface{}{
			"status": "Success",
			"output": fmt.Sprintf("File %s exists", path),
		}, nil
	}
	if os.IsNotExist(err) {
		return map[string]interface{}{
			"status": "Failure",
			"output": fmt.Sprintf("File %s does not exist", path),
		}, nil
	}
	return nil, err
}
