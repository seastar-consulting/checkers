package main

import (
	"fmt"
	"os"

	_ "github.com/seastar-consulting/checkers/checks/all" // Import all checks
	"github.com/seastar-consulting/checkers/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
