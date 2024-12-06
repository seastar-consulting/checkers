package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Config represents the structure of the checks.yaml file
type Config struct {
	Checks []CheckItem `yaml:"checks"`
}

// CheckItem represents individual check configurations
type CheckItem struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Type        string            `yaml:"type"`
	Command     string            `yaml:"command,omitempty"`
	Parameters  map[string]string `yaml:"parameters,omitempty"`
}

// CheckResult represents the result of a single check
type CheckResult struct {
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Status string          `json:"status"`
	Output json.RawMessage `json:"output"`
	Error  string          `json:"error,omitempty"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "checker",
		Short: "A CLI tool to read and process checks from a YAML file",
		Run:   runChecker,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runChecker(cmd *cobra.Command, args []string) {
	// Get current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// Construct path to checks.yaml
	configPath := filepath.Join(dir, "checks.yaml")

	// Read the YAML file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading checks.yaml: %v", err)
	}

	// Parse the YAML content
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	// Create a wait group and results channel for concurrent checks
	var wg sync.WaitGroup
	results := make(chan CheckResult, len(config.Checks))

	// Execute checks concurrently
	for _, check := range config.Checks {
		wg.Add(1)
		go func(check CheckItem) {
			defer wg.Done()
			result := executeCheck(check)
			results <- result
		}(check)
	}

	// Close results channel when all checks are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and print results
	var allResults []CheckResult
	for result := range results {
		allResults = append(allResults, result)
	}

	// Pretty print results as JSON
	prettyResults, err := json.MarshalIndent(allResults, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling results: %v", err)
	}
	fmt.Println(string(prettyResults))
}

func executeCheck(check CheckItem) CheckResult {
	result := CheckResult{
		Name: check.Name,
		Type: check.Type,
	}

	// Handle command type checks
	if check.Type == "command" {
		if check.Command == "" {
			result.Status = "FAILED"
			result.Error = "No command specified"
			return result
		}

		// Execute the command
		cmd := exec.Command("bash", "-c", check.Command)
		var outputBuf, errBuf bytes.Buffer
		cmd.Stdout = &outputBuf
		cmd.Stderr = &errBuf

		err := cmd.Run()
		if err != nil {
			result.Status = "FAILED"
			result.Error = err.Error()

			// If stderr is not empty, use it as output
			if errBuf.Len() > 0 {
				result.Output = json.RawMessage(errBuf.Bytes())
			}
			return result
		}

		// Try to parse output as JSON
		var jsonOutput json.RawMessage
		err = json.Unmarshal(outputBuf.Bytes(), &jsonOutput)
		if err != nil {
			// If not valid JSON, wrap the output as a JSON string
			jsonOutput, _ = json.Marshal(outputBuf.String())
		}

		result.Status = "SUCCESS"
		result.Output = jsonOutput
	}

	return result
}
