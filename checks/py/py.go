package py

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/seastar-consulting/checkers/checks"
	"github.com/seastar-consulting/checkers/types"
	python3 "github.com/seastar-consulting/cpy3"
)

func init() {
	// Register the Python check handler
	checks.Registry["py"] = checks.Check{
		Name:        "Python Function",
		Description: "Execute a Python function in the format py.module.path:function",
		Func:        ExecutePythonCheck,
	}

	// Initialize Python
	python3.Py_Initialize()
}

// goPythonValue converts Go values to Python objects
func goPythonValue(v interface{}) *python3.PyObject {
	switch val := v.(type) {
	case string:
		return python3.PyUnicode_FromString(val)
	case float64:
		return python3.PyFloat_FromDouble(val)
	case bool:
		if val {
			return python3.PyBool_FromLong(1)
		}
		return python3.PyBool_FromLong(0)
	case map[string]interface{}:
		pyDict := python3.PyDict_New()
		for k, v := range val {
			pyKey := python3.PyUnicode_FromString(k)
			pyVal := goPythonValue(v)
			python3.PyDict_SetItem(pyDict, pyKey, pyVal)
			pyKey.DecRef()
			pyVal.DecRef()
		}
		return pyDict
	default:
		// Convert other types to string representation
		return python3.PyUnicode_FromString(fmt.Sprintf("%v", val))
	}
}

// ExecutePythonCheck executes a Python function using cpy3
func ExecutePythonCheck(item types.CheckItem) (types.CheckResult, error) {
	// Split the type into module and function parts (py.module.path:function)
	parts := strings.SplitN(item.Type, ".", 2)
	if len(parts) != 2 {
		return types.CheckResult{}, fmt.Errorf("invalid Python check type format: %s", item.Type)
	}

	// Get the module.path:function part and split into module and function
	funcSpec := parts[1]
	specParts := strings.Split(funcSpec, ":")
	if len(specParts) != 2 {
		return types.CheckResult{}, fmt.Errorf("invalid Python function spec format: %s", funcSpec)
	}
	moduleName := specParts[0]
	funcName := specParts[1]

	// Add current directory to Python path
	_ = python3.PyRun_SimpleString("import sys\nsys.path.append(\"" + filepath.Dir(item.Type) + "\")")

	// Import the module
	module := python3.PyImport_ImportModule(moduleName)
	if module == nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Failure,
			Error:  fmt.Sprintf("could not import module %s", moduleName),
		}, nil
	}
	defer module.DecRef()

	// Get the function
	function := module.GetAttrString(funcName)
	if function == nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Failure,
			Error:  fmt.Sprintf("could not find function %s in module %s", funcName, moduleName),
		}, nil
	}
	defer function.DecRef()

	// Convert parameters to a Python dict
	var params map[string]interface{}
	if len(item.Parameters) > 0 {
		params = make(map[string]interface{})
		for k, v := range item.Parameters {
			params[k] = v
		}
	}

	// Convert items to parameters if present
	if len(item.Items) > 0 {
		params = make(map[string]interface{})
		for _, item := range item.Items {
			for k, v := range item {
				params[k] = v
			}
		}
	}

	// Convert parameters to Python dict
	pyDict := goPythonValue(params)
	defer pyDict.DecRef()

	// Call the function
	result := function.CallFunctionObjArgs(pyDict)
	if result == nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Failure,
			Error:  "Python function call failed",
		}, nil
	}
	defer result.DecRef()

	// Convert result to Go int
	exitCode := python3.PyLong_AsLong(result)

	checkResult := types.CheckResult{
		Name:   item.Name,
		Type:   item.Type,
		Status: types.Success,
		Output: "Python check completed successfully",
	}

	if exitCode != 0 {
		checkResult.Status = types.Failure
		checkResult.Error = fmt.Sprintf("Python check failed with exit code %d", exitCode)
	}

	return checkResult, nil
}
