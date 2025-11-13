package daemon_test

import (
	// DTrack
	"dtrack/daemon"

	// Standard
	"bytes"
	"io"
	"strings"
	"testing"
)

// Ensure input is fully consumed without error
func TestPipe2DevNull(t *testing.T) {
	// Input with some data
	inputString := "data to pipe and discard"
	r := strings.NewReader(inputString)

	// Call the function
	// Function is called directly because the test is in package common.
	daemon.Pipe2DevNull(r)

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
