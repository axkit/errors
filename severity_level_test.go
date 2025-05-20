package errors

import (
	"encoding/json"
	"testing"
)

func TestSeverityLevel_String(t *testing.T) {
	tests := []struct {
		level    SeverityLevel
		expected string
	}{
		{Tiny, stiny},
		{Medium, smedium},
		{Critical, scritical},
		{SeverityLevel(999), sunknown},
	}

	for _, test := range tests {
		result := test.level.String()
		if result != test.expected {
			t.Errorf("Expected %s, but got %s", test.expected, result)
		}
	}
}

func TestSeverityLevel_MarshalJSON(t *testing.T) {
	tests := []struct {
		level    SeverityLevel
		expected []byte
	}{
		{Tiny, tiny},
		{Medium, medium},
		{Critical, critical},
		{SeverityLevel(999), unknown},
	}

	for _, test := range tests {
		result, err := test.level.MarshalJSON()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if string(result) != string(test.expected) {
			t.Errorf("expected %q, but got %q", test.expected, result)
		}
	}
}

func TestSeverityLevel_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    []byte
		expected SeverityLevel
	}{
		{tiny, Tiny},
		{medium, Medium},
		{critical, Critical},
		{unknown, SeverityLevel(0)},
	}

	for _, test := range tests {
		var level SeverityLevel
		err := json.Unmarshal(test.input, &level)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if level != test.expected {
			t.Errorf("expected %q, but got %q", test.expected, level)
		}
	}
}
