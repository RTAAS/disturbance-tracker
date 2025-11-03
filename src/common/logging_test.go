package common

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Helper function to capture stdout
func captureStdout(f func()) string {
	// Save the original stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	// Redirect stdout to the pipe
	os.Stdout = w

	// Execute the function being tested
	f()

	// Close the writing end of the pipe
	w.Close()
	// Restore the original stdout
	os.Stdout = old
	var buf bytes.Buffer
	// Read the captured content
	io.Copy(&buf, r)
	return buf.String()
}

// Helper function to capture stderr
func captureStderr(f func()) string {
	// Save the original stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	// Redirect stderr to the pipe
	os.Stderr = w

	// Execute the function being tested
	f()

	// Close the writing end of the pipe
	w.Close()
	// Restore the original stderr
	os.Stderr = old
	var buf bytes.Buffer
	// Read the captured content
	io.Copy(&buf, r)
	return buf.String()
}

func TestDebug(t *testing.T) {
	testArg := "check"
	expectedFormat := "[DEBUG]\t  Debug message: check\n"

	// 1. Test when Debug_Enabled is TRUE (should print)
	Debug_Enabled = true
	output := captureStdout(func() {
		// FIX: Using a constant string literal for the format to avoid "non-constant format string" error
		Debug("Debug message: %s", testArg)
	})

	if output != expectedFormat {
		t.Errorf("Debug(enabled) failed.\nGot: %q\nWant: %q", output, expectedFormat)
	}

	// 2. Test when Debug_Enabled is FALSE (should NOT print)
	Debug_Enabled = false
	output = captureStdout(func() {
		Debug("This message should not print")
	})

	if output != "" {
		t.Errorf("Debug(disabled) failed. Got output: %q, Want: empty string", output)
	}
}

func TestInfo(t *testing.T) {
	expectedFormat := "[INFO]\t  System initialized\n"

	output := captureStdout(func() {
		// FIX: Using a constant string literal for the format to avoid "non-constant format string" error
		Info("System initialized")
	})

	if output != expectedFormat {
		t.Errorf("Info failed.\nGot: %q\nWant: %q", output, expectedFormat)
	}
}

func TestWarn(t *testing.T) {
	testArg := 512
	expectedFormat := "[WARNING] Memory limit reached: 512MB\n"

	output := captureStderr(func() {
		// FIX: Using a constant string literal for the format to avoid "non-constant format string" error
		Warn("Memory limit reached: %dMB", testArg)
	})

	if output != expectedFormat {
		t.Errorf("Warn failed.\nGot: %q\nWant: %q", output, expectedFormat)
	}
}

func TestDie(t *testing.T) {
	// If the environment variable is set, it means we are in the subprocess.
	if os.Getenv("BE_A_SUBPROCESS") == "1" {
		Die("Test crash reason: %s", "File missing")
		// The os.Exit(1) call in Die prevents this line from being reached.
		return
	}

	// 1. Set up and run the test function as a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestDie")
	// This environment variable tells the subprocess to execute the Die logic
	cmd.Env = append(os.Environ(), "BE_A_SUBPROCESS=1")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr // Capture output from the subprocess's stderr

	err := cmd.Run() // Run the subprocess and wait for it to exit

	// 2. Check the exit status
	if e, ok := err.(*exec.ExitError); ok {
		// Check that the exit code was 1, as expected from Die
		if status, ok := e.Sys().(interface{ ExitCode() int }); ok && status.ExitCode() != 1 {
			t.Errorf("Die exited with status %d, want 1", status.ExitCode())
		}
	} else {
		// If the process returned nil error or a different error, it means os.Exit(1) wasn't called.
		t.Fatalf("Die did not cause the process to exit with an error (expected os.Exit(1)). Got error: %v", err)
	}

	// 3. Check the output on stderr
	expectedStderr := "[CRITICAL] Test crash reason: File missing\n"
	if !strings.Contains(stderr.String(), expectedStderr) {
		t.Errorf("Die failed to print critical message to stderr.\nGot output: %q\nWant substring: %q", stderr.String(), expectedStderr)
	}
}
