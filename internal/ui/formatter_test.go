package ui

import (
	"testing"

	"github.com/seastar-consulting/checkers/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestFormatter_FormatResult(t *testing.T) {
	tests := []struct {
		name     string
		result   types.CheckResult
		checkType string
		isLast   bool
		want     string
	}{
		{
			name: "success result",
			result: types.CheckResult{
				Name:   "test",
				Status: types.Success,
				Output: "test output",
			},
			checkType: "command",
			isLast:  true,
			want:    "└── ✅ test (command)",
		},
		{
			name: "failure result",
			result: types.CheckResult{
				Name:   "test",
				Status: types.Failure,
				Error:  "test error",
			},
			checkType: "command",
			isLast:  true,
			want:    "└── ❌ test (command)\n    ╭────────────╮\n    │ test error │\n    ╰────────────╯",
		},
		{
			name: "error result",
			result: types.CheckResult{
				Name:   "test",
				Status: types.Error,
				Error:  "test error",
			},
			checkType: "command",
			isLast:  true,
			want:    "└── 🟠 test (command)\n    ╭────────────╮\n    │ test error │\n    ╰────────────╯",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(false)
			got := f.FormatResult(tt.result, tt.checkType, tt.isLast)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatter_FormatResults(t *testing.T) {
	tests := []struct {
		name      string
		results   []types.CheckResult
		checkTypes map[string]string
		want      string
	}{
		{
			name: "multiple results same type",
			results: []types.CheckResult{
				{
					Name:   "test1",
					Status: types.Success,
					Output: "test output",
				},
				{
					Name:   "test2",
					Status: types.Failure,
					Error:  "test error",
				},
			},
			checkTypes: map[string]string{
				"test1": "command",
				"test2": "command",
			},
			want: "├── ✅ test1 (command)\n└── ❌ test2 (command)\n    ╭────────────╮\n    │ test error │\n    ╰────────────╯",
		},
		{
			name: "multiple results different types",
			results: []types.CheckResult{
				{
					Name:   "test1",
					Status: types.Success,
					Output: "test output",
				},
				{
					Name:   "test2",
					Status: types.Failure,
					Error:  "test error",
				},
			},
			checkTypes: map[string]string{
				"test1": "command",
				"test2": "script",
			},
			want: "├── command\n├── ✅ test1 (command)\n\n└── script\n└── ❌ test2 (script)\n    ╭────────────╮\n    │ test error │\n    ╰────────────╯",
		},
		{
			name:    "empty results",
			results: []types.CheckResult{},
			checkTypes: map[string]string{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(false)
			got := f.FormatResults(tt.results, tt.checkTypes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatter_FormatResults_DoubleNewline(t *testing.T) {
	// Remove this test as we no longer add double newlines
	t.Skip("Double newline is no longer part of the formatting")
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
