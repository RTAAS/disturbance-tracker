//##
// DTrack Package: Surveilance Monitor
//
// Collects audio+video files and logs any matched audio disturbances.
//##
package daemon

import (
	// DTrack
	"dtrack/log"
	"dtrack/state"
	"dtrack/ffmpeg"
	// Standard
	"os"
	"os/signal"
	"time"
	"io"
)

// Primary post-bootstrap entry point
// Initialize audio segment scanners and begin recording process
func Run() {
	wav_stream, daemon_stream := io.Pipe()
	stop_recording := false

	// Handle interrupt signals
	sig_chan := make(chan os.Signal, 1)
	signal.Notify(sig_chan, os.Interrupt)
	go func() {
		// Wait for first Ctrl+C
		<-sig_chan
		log.Info("SIGTERM: Waiting for current recording to finish.")
		stop_recording = true
		// Wait for second Ctrl+C
		<-sig_chan
		log.Die("Second Ctrl+C received. Terminating immediately.")
	}()

	// Start scanners if any models are defined
	if state.Runtime.Has_Models {
		log.Debug("Initializing segment scanners")
		go start_scanners(wav_stream)
	} else {
		log.Warn("No inspection models configured; only able to record!")
		go Pipe2DevNull(wav_stream)
	}

	// Prepare full ffmpeg command
	record_args := ffmpeg.Recorder_Arguments()
	save_path := state.Runtime.Workspace + "/recordings/"

	// Start main recording loop that sends data to scanners (and mkv recordings)
	for !stop_recording {
		mkv := save_path + time.Now().Format(ffmpeg.SaveName)
		// Verify output directory exists
		if err := os.MkdirAll(save_path, 0755); err != nil {
			log.Die("Failed to make output directory: %s", save_path)
			return
		}

		log.Debug("New ffmpeg process, saving to: %s", mkv)
		args := append(record_args, mkv)
		ffmpeg.ReadStdin(args, daemon_stream, false)

		// Pause to prevent thrashing of physical devices
		time.Sleep(50 * time.Millisecond)
	}
}

// Replicates a stream piped to /dev/null
func Pipe2DevNull(r io.Reader) {
	io.Copy(io.Discard, r)
}

// Initialize all audio segment scanners and process wav_stream data
func start_scanners(wav_stream *io.PipeReader) {
	// Process manager for segment scanners
	scanners := make(map[string]chan audio_segment)
	returned_segments := make(chan audio_segment)

	// Start segment scanner thread for each trained model
	for _, model_name := range state.Runtime.Record_Inspect_Models {
		segment_channel := make(
			chan audio_segment,
			state.Runtime.Record_Inspect_Backlog)
		scanners[model_name] = segment_channel
		go scan_segments(model_name, segment_channel)
	}

	// Stream converter
	go stream_to_segment(wav_stream, returned_segments)

	// Distribute new segments to all scanners
	for {
		select {
		// Collect new segment
		case new_segment, ok := <-returned_segments:
			if !ok {
				log.Die("Stream converter disappeared")
				return
			}
			// Distribute segment to scanners
			for name, scanner := range scanners {
				select {
				// Send segment to individual scanner
				case scanner <- new_segment:
				default:
					log.Warn("Scanner Blocked: %s", name)
				}
			}
		}
	}
}

// Convert an input wav_stream to 1-second audio clips
func stream_to_segment(stream *io.PipeReader, segments chan<- audio_segment) {
	defer close(segments)
	var segment_id uint = 0

	// Start main conversion loop
	for {
		// Allocate a buffer for the audio segment
		segment_data := make([]byte, ffmpeg.BytesPerSecond)
		
		// Block until segment_data is full
		_, err := io.ReadFull(stream, segment_data)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// WAV stream ended; restart fresh loop
			log.Debug("Reading segment was reset")
			continue
		}
		if err != nil {
			log.Die("Unhandled stream read error: %s", err.Error())
		}

		// Add new segment to queue
		log.Trace("New segment accumulated: %d", segment_id)
		segments <- audio_segment {
			count: segment_id,
			data: segment_data,
		}
		segment_id++
	}
}
