package ui

import (
	"strings"
	"testing"

	"github.com/seastar-consulting/checkers/types"
)

func TestFormatter_FormatResult(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		result    types.CheckResult
		wantIcon  string
		wantParts []string
		dontWant  []string
	}{
		{
			name:    "success result - non-verbose",
			verbose: false,
			result: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Success,
				Output: "test output",
			},
			wantIcon:  CheckPassIcon,
			wantParts: []string{"test-check", "test"},
			dontWant:  []string{"test output"},
		},
		{
			name:    "success result - verbose",
			verbose: true,
			result: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Success,
				Output: "test output",
			},
			wantIcon:  CheckPassIcon,
			wantParts: []string{"test-check", "test", "test output"},
		},
		{
			name:    "failure result - non-verbose",
			verbose: false,
			result: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Failure,
				Output: "test failed",
			},
			wantIcon:  CheckFailIcon,
			wantParts: []string{"test-check", "test"},
			dontWant:  []string{"test failed"},
		},
		{
			name:    "error result - multiline non-verbose",
			verbose: false,
			result: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Error,
				Error:  "first line\nsecond line\nthird line",
			},
			wantIcon:  CheckErrorIcon,
			wantParts: []string{"test-check", "test", "first line"},
			dontWant:  []string{"second line", "third line"},
		},
		{
			name:    "error result - multiline verbose",
			verbose: true,
			result: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Error,
				Error:  "first line\nsecond line\nthird line",
			},
			wantIcon:  CheckErrorIcon,
			wantParts: []string{"test-check", "test", "first line", "second line", "third line"},
		},
		{
			name:    "result with no output or error",
			verbose: false,
			result: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Success,
			},
			wantIcon:  CheckPassIcon,
			wantParts: []string{"test-check", "test"},
			dontWant:  []string{"output", "error"},
		},
		{
			name:    "multi-line output with tree structure - verbose",
			verbose: true,
			result: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Success,
				Output: "line1\nline2\nline3",
			},
			wantIcon: CheckPassIcon,
			wantParts: []string{
				"test-check",
				"test",
				"line1",
				"line2",
				"line3",
			},
		},
		{
			name:    "error with multi-line message - verbose",
			verbose: true,
			result: types.CheckResult{
				Name:   "test-check",
				Type:   "test",
				Status: types.Error,
				Error:  "error1\nerror2\nerror3",
			},
			wantIcon: CheckErrorIcon,
			wantParts: []string{
				"test-check",
				"test",
				"error1",
				"error2",
				"error3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.verbose)
			got := f.formatResult(tt.result, true)

			if !strings.Contains(got, tt.wantIcon) {
				t.Errorf("FormatResult() missing icon %q", tt.wantIcon)
			}

			for _, want := range tt.wantParts {
				if !strings.Contains(got, want) {
					t.Errorf("FormatResult() missing %q", want)
				}
			}

			for _, dontWant := range tt.dontWant {
				if strings.Contains(got, dontWant) {
					t.Errorf("FormatResult() contains unwanted %q", dontWant)
				}
			}
		})
	}
}

func TestFormatter_FormatResults(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		results   []types.CheckResult
		wantParts []string
		dontWant  []string
	}{
		{
			name:    "multiple results - non-verbose",
			verbose: false,
			results: []types.CheckResult{
				{
					Name:   "check1",
					Type:   "test",
					Status: types.Success,
					Output: "success output",
				},
				{
					Name:   "check2",
					Type:   "test",
					Status: types.Failure,
					Output: "failure output",
				},
			},
			wantParts: []string{"check1", "check2"},
			dontWant:  []string{"success output", "failure output"},
		},
		{
			name:    "multiple results - verbose",
			verbose: true,
			results: []types.CheckResult{
				{
					Name:   "check1",
					Type:   "test",
					Status: types.Success,
					Output: "success output",
				},
				{
					Name:   "check2",
					Type:   "test",
					Status: types.Failure,
					Output: "failure output",
				},
			},
			wantParts: []string{"check1", "check2", "success output", "failure output"},
		},
		{
			name:      "empty results",
			verbose:   false,
			results:   []types.CheckResult{},
			wantParts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.verbose)
			got := f.FormatResultsPretty(tt.results, types.OutputMetadata{})

			for _, want := range tt.wantParts {
				if !strings.Contains(got, want) {
					t.Errorf("FormatResults() missing %q", want)
				}
			}

			for _, dontWant := range tt.dontWant {
				if strings.Contains(got, dontWant) {
					t.Errorf("FormatResults() contains unwanted %q", dontWant)
				}
			}
		})
	}
}

func TestFormatter_FormatResults_DoubleNewline(t *testing.T) {
	f := NewFormatter(true)
	results := []types.CheckResult{
		{
			Name:   "test1",
			Type:   "test",
			Status: types.Success,
		},
		{
			Name:   "test2",
			Type:   "test",
			Status: types.Success,
		},
	}

	output := f.FormatResultsPretty(results, types.OutputMetadata{})
	if !strings.HasSuffix(output, "\n\n") {
		t.Error("FormatResults output should end with double newline")
	}
}

func TestPrepend(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefix   string
		expected []string
	}{
		{
			name:     "single line",
			input:    "test",
			prefix:   "│ ",
			expected: []string{"│ test"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			prefix:   "│ ",
			expected: []string{"│ line1", "│ line2", "│ line3"},
		},
		{
			name:     "empty string",
			input:    "",
			prefix:   "│ ",
			expected: []string{"│ "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prepend(tt.input, tt.prefix)
			if len(result) != len(tt.expected) {
				t.Errorf("prepend() got %d lines, want %d lines", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("prepend() line %d got = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
