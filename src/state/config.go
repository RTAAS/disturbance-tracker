// Application Configuration
//
// ALL VARIABLES:
// - Define Runtime variables and types:	Application_Configuration
// - Default value of Runtime variables:	Load_Configuration
// - Map JSON Settings to Runtime variables:	Application_Configuration
// - Map Environment vars to Runtime variables:	Environment_Configation_Map
// - Help text showing JSON/EnvVar/Default:	Show_Help
package state

import (
	// DTrack
	"dtrack/log"

	// Standard
	"encoding/json"
	"os"
	"reflect"
	"strconv"
)

// Master object holding loaded configuration data
var Runtime Application_Configuration

// Map json configuration to Runtime
// Defaults set in load_config()
type Application_Configuration struct {
	Workspace              string   `json:"workspace"`
	Workspace_Keep_Temp    bool     `json:"keep_temp"`
	Record_Audio_Device    string   `json:"audio_device"`
	Record_Audio_Options   []string `json:"audio_options"`
	Record_Video_Device    string   `json:"video_device"`
	Record_Video_Options   []string `json:"video_options"`
	Record_Video_Timestamp string   `json:"video_timestamp"`
	Record_Video_Advanced  []string `json:"video_advanced"`
	Record_Inspect_Models  []string `json:"inspect_models"`
	Has_Models             bool
	Record_Inspect_Backlog int      `json:"inspect_backlog"`
	Record_Duration        string   `json:"record_duration"`
	Train_Target           float64  `json:"train_target"`
	Train_Rate             float64  `json:"train_rate"`
	Train_Momentum         float64  `json:"train_momentum"`
	Train_Dropout          float64  `json:"train_dropout"`
}

// Map environment variables to Runtime
var Environment_Configation_Map = map[string]string{
	"DTRACK_WORKSPACE":       "Workspace",
	"DTRACK_KEEP_TEMP":       "Workspace_Keep_Temp",
	"RECORD_AUDIO_DEVICE":    "Record_Audio_Device",
	"RECORD_AUDIO_OPTIONS":   "Record_Audio_Options",
	"RECORD_VIDEO_DEVICE":    "Record_Video_Device",
	"RECORD_VIDEO_OPTIONS":   "Record_Video_Options",
	"RECORD_VIDEO_ADVANCED":  "Record_Video_Advanced",
	"RECORD_INSPECT_MODELS":  "Record_Inspect_Models",
	"RECORD_INSPECT_BACKLOG": "Record_Inspect_Backlog",
	"RECORD_DURATION":        "Record_Duration",
	"TRAIN_TARGET":           "Train_Target",
	"TRAIN_RATE":             "Train_Rate",
	"TRAIN_MOMENTUM":         "Train_Momentum",
	"TRAIN_DROPOUT":          "Train_Dropout",
}

// Loads Runtime configuration data into current state
func Load_Configuration(config_path string) {
	// Default configuration values
	cfg := Application_Configuration{
		Workspace:            "_workspace",
		Workspace_Keep_Temp:  false,
		Record_Audio_Device:  "plughw",
		Record_Audio_Options: []string{"-f", "alsa"},
		Record_Video_Device:  "/dev/video0",
		Record_Video_Options: []string{
			"-f", "v4l2", "-framerate", "15"},
		Record_Video_Timestamp: "drawtext=fontfile=/usr/share/fonts/truetype/freefont/" +
			"FreeMonoBold.ttf:text=%{localtime}:fontcolor=red@0.9:x=7:y=7:fontsize=48",
		Record_Video_Advanced: []string{
			"libx264", "-crf", "23", "-preset", "fast", "-tune", "zerolatency",
			"-maxrate", "3M", "-bufsize", "24M"},
		Record_Duration:        "00:10:00",
		Record_Inspect_Models:  []string{},
		Record_Inspect_Backlog: 5,
		Train_Target:           0.95,
		Train_Rate:             0.001,
		Train_Momentum:         0.9,
		Train_Dropout:          0.2,
	}

	// Check for configuration file
	log.Debug("Loading configuration values from: %s", config_path)
	if _, err := os.Stat(config_path); err != nil {
		log.Info("Configuration file not found; using defaults.")
		Runtime = cfg
		return
	}

	// Load configuration file
	file_data, err := os.ReadFile(config_path)
	if err != nil {
		log.Die("Error opening configuration file; ABORT!")
	}

	// Merge configuration values into cfg
	if err := json.Unmarshal(file_data, &cfg); err != nil {
		log.Die("Failed to parse configuration as JSON; ABORT!")
	}

	// Search for known environment variables
	for env_key, conf_field := range Environment_Configation_Map {
		if env_value := os.Getenv(env_key); env_value != "" {
			log.Debug("Environment variable found: %s", env_key)
			field := reflect.ValueOf(&cfg).Elem().FieldByName(conf_field)
			if !field.IsValid() || !field.CanSet() {
				log.Die("Invalid field: %s", conf_field)
			}

			// Merge environment variables into cfg
			switch field.Kind() {
			case reflect.String:
				field.SetString(env_value)
			case reflect.Int:
				if intVal, err := strconv.Atoi(env_value); err == nil {
					field.SetInt(int64(intVal))
				} else {
					log.Die("%s is not Integer", env_key)
				}
			case reflect.Float64:
				if intVal, err := strconv.Atoi(env_value); err == nil {
					field.SetFloat(float64(intVal))
				} else {
					log.Die("%s is not Float64", env_key)
				}
			case reflect.Bool:
				if boolVal, err := strconv.ParseBool(env_value); err == nil {
					field.SetBool(boolVal)
				} else {
					log.Die("%s is not Boolean", env_key)
				}
			default:
				log.Die("Unexpected field type for %s", conf_field)
			}
		}
	}

	// Helper variables
	cfg.Has_Models = len(cfg.Record_Inspect_Models) > 0

	// Update session variables
	Runtime = cfg
}
