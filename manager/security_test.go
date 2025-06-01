package manager

import (
	"testing"
)

func TestValidatePackageName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		// Valid package names
		{
			name:    "simple package name",
			input:   "vim",
			wantErr: false,
		},
		{
			name:    "package with version",
			input:   "python3.8",
			wantErr: false,
		},
		{
			name:    "package with dash",
			input:   "gcc-9-base",
			wantErr: false,
		},
		{
			name:    "package with underscore",
			input:   "libc6_dev",
			wantErr: false,
		},
		{
			name:    "package with plus",
			input:   "g++",
			wantErr: false,
		},
		{
			name:    "package with architecture",
			input:   "libc6:amd64",
			wantErr: false,
		},
		{
			name:    "package with repo",
			input:   "ppa/package-name",
			wantErr: false,
		},
		{
			name:    "complex valid name",
			input:   "lib32stdc++-9-dev:i386",
			wantErr: false,
		},

		// Invalid package names - command injection attempts
		{
			name:    "semicolon injection",
			input:   "package; rm -rf /",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "pipe injection",
			input:   "package | cat /etc/passwd",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "ampersand injection",
			input:   "package && malicious-command",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "backtick injection",
			input:   "package`evil`",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "dollar sign injection",
			input:   "package$(bad)",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "redirect injection",
			input:   "package > /etc/passwd",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "single quote injection",
			input:   "package'; drop table users; --",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "double quote injection",
			input:   `package"; rm -rf /; "`,
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "backslash injection",
			input:   `package\nmalicious`,
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "space injection",
			input:   "package name with spaces",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "tab injection",
			input:   "package\tmalicious",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "newline injection",
			input:   "package\nmalicious",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "null byte injection",
			input:   "package\x00malicious",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "parenthesis injection",
			input:   "package(malicious)",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "bracket injection",
			input:   "package[malicious]",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "curly brace injection",
			input:   "package{malicious}",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "asterisk injection",
			input:   "package*",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "question mark injection",
			input:   "package?",
			wantErr: true,
			errMsg:  "invalid package name",
		},
		{
			name:    "tilde injection",
			input:   "~package",
			wantErr: true,
			errMsg:  "invalid package name",
		},

		// Edge cases
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "empty",
		},
		{
			name:    "very long name",
			input:   string(make([]byte, 256)),
			wantErr: true,
			errMsg:  "too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePackageName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePackageName(%q) error = %v, want error containing %q", tt.input, err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidatePackageNames(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantErr bool
	}{
		{
			name:    "all valid names",
			input:   []string{"vim", "git", "python3.8"},
			wantErr: false,
		},
		{
			name:    "one invalid name",
			input:   []string{"vim", "git; rm -rf /", "python3.8"},
			wantErr: true,
		},
		{
			name:    "empty slice",
			input:   []string{},
			wantErr: false,
		},
		{
			name:    "empty string in slice",
			input:   []string{"vim", "", "git"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageNames(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePackageNames(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(substr) < len(s) && findSubstring(s, substr)))
}

// Simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
