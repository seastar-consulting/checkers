package main

import (
	"testing"

	"github.com/seastar-consulting/checkers/internal/types"
)

func TestExecuteCheckRaw(t *testing.T) {
	tests := []struct {
		name     string
		check    types.CheckItem
		wantErr  bool
		contains string
	}{
		{
			name: "successful command",
			check: types.CheckItem{
				Name:    "test",
				Command: "echo 'hello'",
			},
			wantErr:  false,
			contains: "hello",
		},
		{
			name: "failing command",
			check: types.CheckItem{
				Name:    "test",
				Command: "exit 1",
			},
			wantErr:  true,
			contains: "error",
		},
		{
			name: "failing command with pipe",
			check: types.CheckItem{
				Name:    "test",
				Command: "exit 1 | echo 'hello'",
			},
			wantErr:  true,
			contains: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executeCheckRaw(tt.check)
			if tt.wantErr && result["status"] != "error" {
				t.Errorf("executeCheckRaw() expected error status, got %v", result["status"])
			}
			if !tt.wantErr && result["status"] != "success" {
				t.Errorf("executeCheckRaw() expected success status, got %v", result["status"])
			}
		})
	}
}

func TestProcessOutput(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  types.CheckResult
	}{
		{
			name: "success case",
			input: map[string]interface{}{
				"name":   "test",
				"status": "success",
				"output": "test output",
			},
			want: types.CheckResult{
				Name:   "test",
				Status: types.Success,
				Output: "test output",
			},
		},
		{
			name: "error case",
			input: map[string]interface{}{
				"name":   "test",
				"status": "error",
				"error":  "failed",
			},
			want: types.CheckResult{
				Name:   "test",
				Status: types.Error,
				Error:  "failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processOutput(tt.input)
			if got.Status != tt.want.Status {
				t.Errorf("processOutput() status = %v, want %v", got.Status, tt.want.Status)
			}
		})
	}
}
