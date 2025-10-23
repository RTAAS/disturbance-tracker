package state

import (
	"os"
	"testing"
)

func TestLoad_Configuration_WithTempFile(t *testing.T) {
	// Sample demo JSON configuration
	demoJSON := `{
		"workspace": "/tmp/demo_workspace",
		"record_audio_device": "demo_mic",
		"record_audio_options": ["-f", "alsa"],
		"record_video_device": "/dev/video9",
		"record_video_options": ["-f", "v4l2"],
		"record_video_advanced": ["-filter_complex", "[0:v]hflip"],
		"record_inspect_models": ["model1", "model2"],
		"record_inspect_backlog": 3,
		"record_inspect_segment": 10,
		"record_duration": "00:05:00",
		"record_compression": "fast",
		"train_target": 0.85,
		"train_rate": 0.005,
		"train_momentum": 0.8,
		"train_dropout": 0.3
	}`

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // delete after test

	// Write demo JSON to temp file
	if _, err := tmpFile.Write([]byte(demoJSON)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load configuration from the temp file
	Load_Configuration(tmpFile.Name())

	cfg := Runtime

	// Verify some fields
	if cfg.Workspace != "/tmp/demo_workspace" {
		t.Errorf("Expected workspace '/tmp/demo_workspace', got '%s'", cfg.Workspace)
	}
	if cfg.Record_Audio_Device != "demo_mic" {
		t.Errorf("Expected Record_Audio_Device 'demo_mic', got '%s'", cfg.Record_Audio_Device)
	}
	if len(cfg.Record_Inspect_Models) != 2 {
		t.Errorf("Expected 2 inspect models, got %d", len(cfg.Record_Inspect_Models))
	}
	if cfg.Train_Target != 0.85 {
		t.Errorf("Expected Train_Target 0.85, got %f", cfg.Train_Target)
	}
}
