//##
// Audio Disturbance Tracker (DTrack)
//
// License:   AGPL-3
// Copyright: 2024-2026, Michael Lustfield (MTecknology)
// Authors:   See history with "git log" or "git blame"
//##
package main

import (
	// Bootstrap
	"dtrack/common"
	"dtrack/state"
	// Actions
	"dtrack/daemon"
	"dtrack/review"
	"dtrack/model"
)

func main() {
	// Bootstrap
	parse_flags()
	common.Debug_Enabled = *app_verbose || *app_trace
	common.Trace_Enabled = *app_trace
	state.Load_Configuration(*app_config_path)

	// Kickoff
	action_map := map[string]func() {
		"monitor": daemon.Run,
		"record":  daemon.Run, // Alias
		"review":  review.Start,
		"train":   model.Train,
	}
	action_map[*app_action]()

	// Post-processing
	common.Clean_Workspace()
}
