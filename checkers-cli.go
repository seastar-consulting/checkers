package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
	Parameters  map[string]string `yaml:"parameters,omitempty"`
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

	// Print out the parsed configuration
	fmt.Println("Checks found in configuration:")
	for i, check := range config.Checks {
		fmt.Printf("%d. Name: %s\n", i+1, check.Name)
		if check.Description != "" {
			fmt.Printf("   Description: %s\n", check.Description)
		}
		fmt.Printf("   Type: %s\n", check.Type)
		if len(check.Parameters) > 0 {
			fmt.Println("   Parameters:")
			for key, value := range check.Parameters {
				fmt.Printf("     - %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}
}
