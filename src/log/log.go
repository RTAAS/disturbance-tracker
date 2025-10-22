package log

import (
	// Standard
	"fmt"
	"os"
)

// Toggles to change logging verbosity
var Debug_Enabled bool = false
var Trace_Enabled bool = false

// Only print a message to stdout if verbosity is enabled
func Trace(format string, a ...interface{}) {
	if Trace_Enabled {
		message := fmt.Sprintf(format, a...)
		fmt.Println("[TRACE]\t", message)
	}
}

// Only print a message to stdout if verbosity is enabled
func Debug(format string, a ...interface{}) {
	if Debug_Enabled {
		message := fmt.Sprintf(format, a...)
		fmt.Println("[DEBUG]\t", message)
	}
}

// Print a message to stdout
func Info(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	fmt.Println("[INFO]\t", message)
}

// Print a message to stderr
func Warn(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, "[WARN]\t", message)
}

// Print a message to stderr, run cleanup, and exit
func Die(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, "[ERROR]\t", message)
	Clean_Workspace()
	os.Exit(1)
}

// Clean up any partial operations found in workspace
func Clean_Workspace() {
	// No-Op
}
