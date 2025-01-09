package errors

import (
	"errors"
	"testing"
)

func TestCheckError(t *testing.T) {
	tests := []struct {
		name     string
		checkName string
		err      error
		want     string
	}{
		{
			name:     "basic error",
			checkName: "test-check",
			err:      errors.New("something failed"),
			want:     `check "test-check" failed: something failed`,
		},
		{
			name:     "nil error",
			checkName: "test-check",
			err:      nil,
			want:     `check "test-check" failed: <nil>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewCheckError(tt.checkName, tt.err)
			if got := e.Error(); got != tt.want {
				t.Errorf("CheckError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name  string
		field string
		err   error
		want  string
	}{
		{
			name:  "basic error",
			field: "config-field",
			err:   errors.New("invalid value"),
			want:  `config error in field "config-field": invalid value`,
		},
		{
			name:  "nil error",
			field: "config-field",
			err:   nil,
			want:  `config error in field "config-field": <nil>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewConfigError(tt.field, tt.err)
			if got := e.Error(); got != tt.want {
				t.Errorf("ConfigError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
