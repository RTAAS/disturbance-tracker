package ffmpeg_test

import (
	// DTrack
	"dtrack/state"
	"dtrack/ffmpeg"

	// Standard
	"reflect"
	"testing"
)

// Mock the state package's Runtime for testing Recorder_Arguments
func setupMockState() {
	state.Runtime = state.Application_Configuration{
		Record_Duration:       "10",
		Record_Audio_Options:  []string{"-f", "alsa"},
		Record_Audio_Device:   "hw:0",
		Record_Video_Options:  []string{"-f", "v4l2", "-framerate", "30"},
		Record_Video_Device:   "/dev/video0",
		Record_Video_Timestamp: "drawtext=text='%{localtime}':fontcolor=white:fontsize=24:x=10:y=10",
		Has_Models:            true,
		// We need to set Record_Video_Advanced to a known value for comparison
		Record_Video_Advanced: []string{"libx264", "-preset", "ultrafast"}, 
	}
}

// Checks if the arguments for extraction are correctly formed.
func TestExtractArguments(t *testing.T) {
	t.Parallel()
	infile := "test.mkv"
	outdir := "/tmp/output"
	
	// Expected arguments array
	expected := []string{
		// basic-options input-mkv
		"-y", "-loglevel", "warning", "-nostdin", "-nostats", "-i", infile,
		// wav-to-stdout
		"-map", "0:a:0", "-f", "s16le", "-ar", "48000", "-ac", "1", "-",
		// output-wav
		"-f", "segment", "-segment_time", "1", "-reset_timestamps", "1", outdir + "/%d.wav",
		// output-images
		"-map", "0:v:0", "-vf", "fps=1,scale=1536:864", "-start_number", "0", outdir + "/%d.jpg",
	}

	actual := ffmpeg.Extract_Arguments(infile, outdir)
	
	// Use reflect.DeepEqual for a deep comparison of the string slices
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Extract_Arguments returned incorrect arguments.\nExpected: %v\nActual:   %v", expected, actual)
	}
}

// Checks if the arguments for recording are correctly formed based on mocked state.
func TestRecorderArguments(t *testing.T) {
	t.Parallel()
	// 1. Setup mock state
	setupMockState()

	// 2. Define expected arguments
	expected := []string{
		// basic-options
		"-y", "-loglevel", "warning", "-nostdin", "-nostats", "-guess_layout_max", "1",
		// audio-options
		"-t", "10", "-f", "alsa",
		// audio-device
		"-i", "hw:0",
		// video-options
		"-t", "10", "-f", "v4l2", "-framerate", "30",
		// video-device
		"-i", "/dev/video0",
		// wav-to-stdout (Has_Models is true)
		"-map", "0:a", "-c:a", "pcm_s16le", "-ar", "48000", "-ac", "1", "-f", "wav", "-",
		// wav&vid-to-mkv
		"-filter_complex", "[1:v]drawtext=text='%{localtime}':fontcolor=white:fontsize=24:x=10:y=10[dtstamp]",
		"-map", "0:a", "-map", "[dtstamp]", "-c:a", "pcm_s16le",
		"-ar", "48000", "-ac", "1", "-c:v",
		"libx264", "-preset", "ultrafast", // Added from mocked Record_Video_Advanced
	}

	// 3. Get actual arguments
	actual := ffmpeg.Recorder_Arguments()

	// 4. Verify using reflect.DeepEqual
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Recorder_Arguments returned incorrect arguments.\nExpected: %v\nActual:   %v", expected, actual)
	}
}
