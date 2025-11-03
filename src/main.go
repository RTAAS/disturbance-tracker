// ##
// Audio Disturbance Tracker (DTrack)
//
// License:	AGPL-3
// Copyright:	2024-2026, Michael Lustfield (MTecknology)
// Authors:	See history with "git log" or "git blame"
// ##
package main

import (
	// Bootstrap
	"dtrack/log"
	"dtrack/state"

	// Actions
	"dtrack/daemon"
	"dtrack/model"
	"dtrack/review"
)

func main() {
	// Bootstrap
	parse_flags()
	log.Debug_Enabled = *app_verbose || *app_trace
	log.Trace_Enabled = *app_trace
	state.Load_Configuration(*app_config_path)
	if *app_keep_temp {
		state.Runtime.Workspace_Keep_Temp = true
	}
	defer Clean_Workspace()

	// Kickoff
	action_map := map[string]func(){
		"monitor": daemon.Run,
		"record":  daemon.Run, // Alias
		"review":  review.Launch,
		"train":   model.Train,
	}
	action_map[*app_action]()

}

// Post-processing
func Clean_Workspace() {
	// No-Op
}
