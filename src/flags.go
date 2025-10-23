// Command Line Argument (a.k.a. Golang Flags)
package main

import (
	// DTrack
	"dtrack/log"
	"dtrack/state"

	// Standard
	"flag"
	"fmt"
)

// Application flags
var (
	app_action = flag.String(
		"a", "<none>",
		"Application action (See Actions, above)")
	app_config_path = flag.String(
		"c", "./config.json",
		"Path to configuration file")
	app_verbose = flag.Bool(
		"v", false,
		"Enable verbose logging.")
	app_trace = flag.Bool(
		"V", false,
		"Like -v, but more.")
)

// Parse command-line arguments (flags)
func parse_flags() {
	flag.Usage = show_help
	flag.Parse()

	// Safety checks
	okay_actions := []string{"monitor", "review", "train", "record"}
	if !In_List(*app_action, okay_actions) {
		show_help()
		log.Die("Unexpected Action: %s", *app_action)
	}
}

// Show basic usage information
func show_help() {
	fmt.Println("Usage:\n  dtracker [-h] -a <action> [options]")
	fmt.Println("\nActions:") // copy: okay_actions
	fmt.Println("  monitor\tCollect recordings and automatically review")
	fmt.Println("  review\tManually review collected logs")
	fmt.Println("  train\t\tTrain a new AI Model")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	state.Show_Help() // config.go
	fmt.Println("\nExamples:")
	fmt.Println("  DTRACK_RECORD_DURATION=00:05:00  dtrack -a monitor")
	fmt.Println("  dtrack -a review")
}

// Returns true if a search string is present in a list of slices
func In_List(needle string, haystack []string) bool {
	// Iterate through each element in the slice
	for _, element := range haystack {
		// Check if the current element matches
		if element == needle {
			return true
		}
	}
	// No match found
	return false
}
