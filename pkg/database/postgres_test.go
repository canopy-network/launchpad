package database

import (
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no null bytes",
			input:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			expected: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		},
		{
			name:     "single null byte in middle",
			input:    "Mozilla/5.0\x00Windows",
			expected: "Mozilla/5.0Windows",
		},
		{
			name:     "multiple null bytes",
			input:    "Test\x00String\x00With\x00Nulls",
			expected: "TestStringWithNulls",
		},
		{
			name:     "null byte at start",
			input:    "\x00StartWithNull",
			expected: "StartWithNull",
		},
		{
			name:     "null byte at end",
			input:    "EndWithNull\x00",
			expected: "EndWithNull",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only null bytes",
			input:    "\x00\x00\x00",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNullString(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		result := NullString(nil)
		if result.Valid {
			t.Error("Expected Valid to be false for nil pointer")
		}
	})

	t.Run("valid string without null bytes", func(t *testing.T) {
		input := "test string"
		result := NullString(&input)
		if !result.Valid {
			t.Error("Expected Valid to be true")
		}
		if result.String != input {
			t.Errorf("Expected %q, got %q", input, result.String)
		}
	})

	t.Run("string with null bytes", func(t *testing.T) {
		input := "test\x00string\x00with\x00nulls"
		expected := "teststringwithnulls"
		result := NullString(&input)
		if !result.Valid {
			t.Error("Expected Valid to be true")
		}
		if result.String != expected {
			t.Errorf("Expected %q, got %q", expected, result.String)
		}
	})
}
