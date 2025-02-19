package py

import (
	"fmt"
	"os"
	"strings"

	"github.com/seastar-consulting/checkers/checks"
	"github.com/seastar-consulting/checkers/types"
	python3 "github.com/seastar-consulting/cpy3"
)

func init() {
	// Register the Python check handler
	checks.Register("py", "Execute a Python function in the format py.module.path:function", ExecutePythonCheck)

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
	cwd, err := os.Getwd()
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Failure,
			Error:  fmt.Sprintf("could not get current working directory: %v", err),
		}, nil
	}
	_ = python3.PyRun_SimpleString("import sys\nsys.path.append(\"" + cwd + "\")")

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

	// Convert Python dict to Go map
	checkResult := types.CheckResult{
		Name: item.Name,
		Type: item.Type,
	}

	// Get status
	statusObj := python3.PyDict_GetItemString(result, "status")
	if statusObj != nil {
		status := python3.PyUnicode_AsUTF8(statusObj)
		switch status {
		case "success":
			checkResult.Status = types.Success
		case "failure":
			checkResult.Status = types.Failure
		default:
			checkResult.Status = types.Error
		}
	} else {
		checkResult.Status = types.Error
		checkResult.Error = "no status returned from Python function"
		return checkResult, nil
	}

	// Get output if present
	if outputObj := python3.PyDict_GetItemString(result, "output"); outputObj != nil {
		checkResult.Output = python3.PyUnicode_AsUTF8(outputObj)
	}

	// Get error if present
	if errorObj := python3.PyDict_GetItemString(result, "error"); errorObj != nil {
		checkResult.Error = python3.PyUnicode_AsUTF8(errorObj)
	}

	return checkResult, nil
}
