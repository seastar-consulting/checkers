package checks

import "github.com/seastar-consulting/checkers/types"

// CheckFunc is a function that implements a check
type CheckFunc func(item types.CheckItem) (types.CheckResult, error)

// Check represents a registered check
type Check struct {
	Name        string
	Description string
	Func        CheckFunc
}
