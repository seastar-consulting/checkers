package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/seastar-consulting/checkers/types"
)

func TestFormatter_FormatResultsHTML(t *testing.T) {
	// Create a formatter
	formatter := NewFormatter(true)

	// Create test results with different statuses
	results := []types.CheckResult{
		{
			Name:   "Success Test",
			Status: types.Success,
			Type:   "test.success",
			Output: "This check passed successfully",
		},
		{
			Name:   "Warning Test",
			Status: types.Warning,
			Type:   "test.warning",
			Output: "This check produced a warning",
		},
		{
			Name:   "Error Test",
			Status: types.Error,
			Type:   "test.error",
			Error:  "This check failed with an error",
		},
		{
			Name:   "Failure Test",
			Status: types.Failure,
			Type:   "test.failure",
			Error:  "This check failed",
		},
	}

	// Create metadata
	metadata := types.OutputMetadata{
		DateTime: "2025-03-05T12:00:00Z", // Fixed time for consistent test results
		Version:  "1.0.0-test",
		OS:       "test-os/test-arch",
	}

	// Format results as HTML
	html := formatter.FormatResultsHTML(results, metadata)

	// Test that the HTML contains expected elements
	expectedElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"<head",
		"<body",
		"Success Test",
		"Warning Test",
		"Error Test",
		"Failure Test",
		"test.success",
		"test.warning",
		"test.error",
		"test.failure",
		"This check passed successfully",
		"This check produced a warning",
		"This check failed with an error",
		"This check failed",
		"1.0.0-test", // Version from metadata
		"test-os/test-arch", // OS from metadata
	}

	for _, expected := range expectedElements {
		if !strings.Contains(html, expected) {
			t.Errorf("FormatResultsHTML() output missing expected element: %q", expected)
		}
	}

	// Test that results are grouped correctly - the template uses group-header class
	if !strings.Contains(html, "class=\"group-header\"") {
		t.Errorf("FormatResultsHTML() output missing expected group header class")
	}

	// Test that success/failure classes are applied correctly
	// The template uses the status directly as a class name (lowercase)
	statusClasses := []string{
		"class=\"check success\"",
		"class=\"check warning\"",
		"class=\"check error\"",
		"class=\"check failure\"",
	}

	for _, class := range statusClasses {
		if !strings.Contains(html, class) {
			t.Errorf("FormatResultsHTML() output missing expected class: %q", class)
		}
	}
}

func TestFormatter_FormatResultsHTML_EmptyResults(t *testing.T) {
	// Test with empty results
	formatter := NewFormatter(true)
	metadata := types.OutputMetadata{
		DateTime: time.Now().Format(time.RFC3339),
		Version:  "1.0.0",
		OS:       "test-os/test-arch",
	}
	
	html := formatter.FormatResultsHTML([]types.CheckResult{}, metadata)
	
	// Should still produce valid HTML
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("FormatResultsHTML() with empty results should still produce valid HTML")
	}
	
	// Should include metadata even with empty results
	if !strings.Contains(html, "1.0.0") || !strings.Contains(html, "test-os/test-arch") {
		t.Errorf("FormatResultsHTML() with empty results should still include metadata")
	}
}
