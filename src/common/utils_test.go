package common

import (
	"bytes"
	"io"
	"strings"
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
			haystack: []string {"monitor", "review", "train"},
			expected: true,
		},
		{
			name:     "Not present in list",
			needle:   "wrongaction",
			haystack: []string {"monitor", "review", "train"},
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
			haystack: []string {"monitor", "review", "train"},
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

func TestClean_Workspace(t *testing.T) {
	// No-Op
}

// Ensure input is fully consumed without error
func TestPipe2DevNull(t *testing.T) {
	// Input with some data
	inputString := "data to pipe and discard"
	r := strings.NewReader(inputString)

	// Call the function
	// Function is called directly because the test is in package common.
	Pipe2DevNull(r)

	// Verify the reader is fully consumed by trying to read again.
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)

	if err != nil && err != io.EOF {
		t.Fatalf("Unexpected error after Pipe2DevNull: %v", err)
	}

	if n > 0 {
		t.Errorf("Reader was not fully consumed. Read %d more bytes: %q", n, buf.String())
	}
}
