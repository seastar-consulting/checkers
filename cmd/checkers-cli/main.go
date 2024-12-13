package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

	// Get current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// Construct path to checks.yaml
	configPath := filepath.Join(dir, configFile)

	if verbose {
		log.Printf("Using config file: %s", configPath)
	}

	// Read the YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading checks.yaml: %v", err)
	}

	// Parse the YAML content
	var config types.Config // Use imported type
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	// Create a wait group and results channel for concurrent checks
	var wg sync.WaitGroup
	results := make(chan types.CheckResult, len(config.Checks))

	// Execute checks concurrently
	for _, check := range config.Checks {
		wg.Add(1)
		go func(check types.CheckItem) {
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
	var allResults []types.CheckResult
	for result := range results {
		allResults = append(allResults, result)
	}

	// Pretty print results as JSON
	// prettyResults, err := json.MarshalIndent(allResults, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Error marshaling results: %v", err)
	// }
	// fmt.Println(string(prettyResults))

	// Process results
	processResults(allResults)
}

// Add new function to format output
func formatCheckResult(result types.CheckResult) string {
	var icon string
	statusIconMap := map[types.CheckStatus]string{
		types.Success: checkPassIcon,
		types.Error:   checkErrorIcon,
		types.Failure: checkFailIcon,
	}

	icon, ok := statusIconMap[result.Status]
	if !ok {
		icon = checkErrorIcon
	}
	return fmt.Sprintf("[%s] %s", icon, result.Name)
}

// Update results processing in runChecker
func processResults(results []types.CheckResult) {
	for _, result := range results {
		fmt.Println(formatCheckResult(result))
		if result.Status != types.Success {
			fmt.Printf("   Error: %v\n", result.Error)
			fmt.Printf("   Output: %s\n", result.Output)
		}
	}
}

func executeCheck(check types.CheckItem) types.CheckResult {
	result := types.CheckResult{
		Name: check.Name,
	}

	if check.Type == "command" {
		if check.Command == "" {
			result.Status = types.Error
			return result
		}

		cmd := exec.Command("bash", "-c", check.Command)
		var outputBuf, errBuf bytes.Buffer
		cmd.Stdout = &outputBuf
		cmd.Stderr = &errBuf

		err := cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.Status = types.Error
				result.Error = fmt.Sprintf("command failed with exit code %d", exitErr.ExitCode())
			}

			if errBuf.Len() > 0 {
				result.Error = errBuf.String()
			}
			return result
		}

		// Validate and handle successful output
		var output map[string]interface{}
		if err := json.Unmarshal(outputBuf.Bytes(), &output); err == nil {
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
			result.Output = outputBuf.String()
		} else {
			result.Output = outputBuf.String()
			result.Status = types.Error
		}

		return result
	}

	return result
}
