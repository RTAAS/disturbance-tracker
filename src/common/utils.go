package common

import (
	"io"
)

// Clean up any partial operations found in workspace
func Clean_Workspace() {
	// No-Op
}

// Replicates a stream piped to /dev/null
func Pipe2DevNull(r io.Reader) {
	io.Copy(io.Discard, r)
}

// Returns true if a search string is present in a list of slices
func In_List(needle string, haystack []string) bool {
    // Iterate through each element in the slice
    for _, element := range haystack {
        // Check if the current element matches the needle
        if element == needle {
            return true
        }
    }
    // No match found
    return false
}
