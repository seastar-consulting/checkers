package py

/*
#cgo linux CFLAGS: -I/usr/include/python3.8
#cgo linux LDFLAGS: -lpython3.8
#cgo darwin CFLAGS: -I/opt/homebrew/opt/python@3.13/Frameworks/Python.framework/Versions/3.13/include/python3.13
#cgo darwin LDFLAGS: -L/opt/homebrew/opt/python@3.13/Frameworks/Python.framework/Versions/3.13/lib -lpython3.13
#cgo windows CFLAGS: -IC:/Python38/include
#cgo windows LDFLAGS: -LC:/Python38/libs -lpython38

#include <Python.h>
*/
import "C"

// This file is needed to specify the build tags and linking flags for Python integration.
// It supports multiple platforms:
// - Linux: Uses system Python 3.8
// - macOS: Uses Homebrew Python 3.13 (Apple Silicon)
// - Windows: Uses Python 3.8 installed in C:/Python38
//
// To use a different Python version or installation path:
// 1. Modify the CFLAGS to point to your Python include directory
// 2. Modify the LDFLAGS to point to your Python library directory and use the correct library name
//
// Example for a different Python version on macOS with Homebrew:
// #cgo darwin CFLAGS: -I/opt/homebrew/opt/python@3.9/Frameworks/Python.framework/Versions/3.9/include/python3.9
// #cgo darwin LDFLAGS: -L/opt/homebrew/opt/python@3.9/Frameworks/Python.framework/Versions/3.9/lib -lpython3.9
