package main

import (
	"testing"
)

func TestIn_List(t *testing.T) {
	tests := []struct {
		name     string
		needle   string
		haystack []string
		expected bool
	}{
		{
			name:     "Present in list",
			needle:   "monitor",
			haystack: []string{"monitor", "review", "train"},
			expected: true,
		},
		{
			name:     "Not present in list",
			needle:   "wrongaction",
			haystack: []string{"monitor", "review", "train"},
			expected: false,
		},
		{
			name:     "Empty list",
			needle:   "oops",
			haystack: []string{},
			expected: false,
		},
		{
			name:     "Case sensitive (Not found)",
			needle:   "MONITOR",
			haystack: []string{"monitor", "review", "train"},
			expected: false,
		},
		{
			name:     "Single element match",
			needle:   "only",
			haystack: []string{"only"},
			expected: true,
		},
		{
			name:     "Single element no match",
			needle:   "wrong",
			haystack: []string{"only"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Functions are called directly because the test is in package common.
			actual := In_List(tt.needle, tt.haystack)
			if actual != tt.expected {
				t.Errorf("In_List(%q, %v) got %v, want %v", tt.needle, tt.haystack, actual, tt.expected)
			}
		})
	}
}
