package events

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "PT30S"},
		{60 * time.Second, "PT1M"},
		{90 * time.Second, "PT1M30S"},
		{5 * time.Minute, "PT5M"},
		{65 * time.Second, "PT1M5S"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
		}
	}
}

func TestSplitSpaceSeparated(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"mqtt", []string{"mqtt"}},
		{"mqtt mqtts", []string{"mqtt", "mqtts"}},
		{"  mqtt   mqtts  ", []string{"mqtt", "mqtts"}},
		{"a b c", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		result := splitSpaceSeparated(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitSpaceSeparated(%q) returned %d items, expected %d", tt.input, len(result), len(tt.expected))

			continue
		}

		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("splitSpaceSeparated(%q)[%d] = %q, expected %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}
