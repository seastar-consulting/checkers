package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/seastar-consulting/checkers/internal/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// defaultConfigFile is the default path to the configuration file
const defaultConfigFile = "checks.yaml"

// Add these constants at the top
const (
	checkPassIcon  = "✅"
	checkFailIcon  = "❌"
	checkErrorIcon = "🟠"
)

// Add new struct for raw output
type rawOutput struct {
	name   string
	output string
}

func main() {
	var (
		configFile string
		verbose    bool
	)

	rootCmd := &cobra.Command{
		Use:   "checker",
		Short: "A CLI tool to read and process checks from a YAML file",
		Run:   runChecker,
	}

	// Add flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", defaultConfigFile, "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

func runChecker(cmd *cobra.Command, args []string) {
	configFile, _ := cmd.Flags().GetString("config")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		log.Printf("Using config file: %s", configFile)
	}

	// Read the YAML file
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading checks.yaml: %v", err)
	}

	// Parse the YAML content
	var config types.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	// Create a wait group and results channel for concurrent checks
	var wg sync.WaitGroup
	results := make(chan map[string]interface{}, len(config.Checks))

	// Execute checks concurrently
	for _, check := range config.Checks {
		wg.Add(1)
		go func(check types.CheckItem) {
			defer wg.Done()
			output := executeCheckRaw(check)
			results <- output
		}(check)
	}

	// Close results channel when all checks are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect outputs and process them into CheckResults
	var checkResults []types.CheckResult
	for output := range results {
		result := processOutput(output)
		checkResults = append(checkResults, result)
	}

	// Print results
	processResults(checkResults)
}

func executeCheckRaw(check types.CheckItem) map[string]interface{} {
	wrappedCmd := fmt.Sprintf("set -eo pipefail; %s", check.Command)
	cmd := exec.Command("bash", "-c", wrappedCmd)

	var outputBuf, errBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &errBuf

	result := map[string]interface{}{
		"name": check.Name,
	}

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			result["status"] = "error"
			result["error"] = fmt.Sprintf("command failed with exit code %d", exitCode)
			result["output"] = errBuf.String()
			result["exitCode"] = exitCode
			return result
		}
		result["status"] = "error"
		result["error"] = err.Error()
		return result
	}

	// Try to parse output as JSON first
	var outputMap map[string]interface{}
	if err := json.Unmarshal(outputBuf.Bytes(), &outputMap); err == nil {
		// Merge command output with result
		for k, v := range outputMap {
			result[k] = v
		}
		if _, ok := result["status"]; !ok {
			result["status"] = "success"
		}
	} else {
		// Raw output
		result["status"] = "success"
		result["output"] = outputBuf.String()
	}

	return result
}

func processOutput(output map[string]interface{}) types.CheckResult {
	var result types.CheckResult

	// Get status from output map
	if status, ok := output["status"].(string); ok {
		switch status {
		case "success":
			result.Status = types.Success
		case "failure":
			result.Status = types.Failure
		case "warning":
			result.Status = types.Warning
		default:
			result.Status = types.Error
		}
	} else {
		result.Status = types.Error
	}

	// Set name from output map
	if name, ok := output["name"].(string); ok {
		result.Name = name
	}

	// Set error if present
	if err, ok := output["error"].(string); ok {
		result.Error = err
	}

	// Set raw output
	if rawOutput, ok := output["output"].(string); ok {
		result.Output = rawOutput
	}

	return result
}

func processResults(results []types.CheckResult) {
	for _, result := range results {
		fmt.Println(formatCheckResult(result))
		if result.Status != types.Success {
			fmt.Printf("   Error: %v\n", result.Error)
			fmt.Printf("   Output: %s\n", result.Output)
		}
	}
}

func formatCheckResult(result types.CheckResult) string {
	var icon string
	switch result.Status {
	case types.Success:
		icon = checkPassIcon
	case types.Failure:
		icon = checkFailIcon
	case types.Warning:
		icon = checkErrorIcon
	default:
		icon = checkErrorIcon
	}
	return fmt.Sprintf("[%s] %s", icon, result.Name)
}
