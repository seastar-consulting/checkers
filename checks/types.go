package checks

// CheckFunc is a function that performs a check
type CheckFunc func(params map[string]interface{}) (map[string]interface{}, error)

// Check represents a registered check
type Check struct {
	Name        string
	Description string
	Func        CheckFunc
}
