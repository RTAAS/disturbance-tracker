package ffmpeg

import (
	// DTrack
	. "dtrack/common"
	"dtrack/state"
	// Standard
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// One-second segment of audio from pcm_s16le (segment_size)
//   Bytes Per Second = Sample Rate * Channels * (Bits Per Sample / 8)
//   96000            = -ac 48000   * -c 1     * (16/8)
const BytesPerSecond int = 96000

// MKV Filename:  YYYY-MM-DD_HHmmss
const SaveName = "2006-01-02_150405.mkv"

// Run ffmpeg command, returning stdout to IO stream
func ReadStdin(arguments []string, stdout *io.PipeWriter) {
	ffmpeg := exec.Command("ffmpeg", arguments...)
	ffmpeg.Stderr = os.Stderr
	ffmpeg.Stdout = stdout

	// Use separate process group to avoid SIGTERM collisions
	ffmpeg.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start ffmpeg process
	if ffmpeg.Start() != nil {
		Die("Failed to intialize ffmpeg")
	}
	if ffmpeg.Wait() != nil {
		Warn("ffmpeg finished with errors")
		// Extra pause for potential device thrashing
		time.Sleep(1 * time.Second)
	}
}

// Return list of arguments for ffmpeg that saves A/V to MKV and Audio to Stream.
// ffmpeg [basic-options] \
//   [audio-options] [audio-device] \
//   [video-options] [video-device] \
//   [output-wav] [to-stdout] \
//   [output-wav&vid] [to-mkv] [MISSING:filename]
func Recorder_Arguments() []string {
	// 5+2+_+2+2+_+2+11+_+14 = 38 (+vars)
	arg_count := 38 +
		len(state.Runtime.Record_Audio_Options) +
		len(state.Runtime.Record_Video_Options) +
		len(state.Runtime.Record_Video_Advanced)
	if !state.Runtime.Has_Models {
		arg_count -= 11
	}
	// Base arguments for ffmpeg (without filename)
	args := make([]string, 0, arg_count)

	// basic-options  +5
	args = append(args, "-y", "-loglevel", "fatal", "-nostdin", "-nostats")
	
	// audio-options  +2 +X
	args = append(args, "-t", state.Runtime.Record_Duration)
	args = append(args, state.Runtime.Record_Audio_Options...)
	// audio-device   +2
	args = append(args, "-i", state.Runtime.Record_Audio_Device)

	// video-options  +2 +X
	args = append(args, "-t", state.Runtime.Record_Duration)
	args = append(args, state.Runtime.Record_Video_Options...)
	// video-device   +2
	args = append(args, "-i", state.Runtime.Record_Video_Device)

	// wav-to-stdout  +11
	if state.Runtime.Has_Models {
		args = append(args,
			"-map", "0:a", "-c:a", "pcm_s16le",
			"-ar", "48000", "-ac", "1", "-f", "wav", "-")
	}
	// wav&vid-to-mkv +X +14
	args = append(args, state.Runtime.Record_Video_Advanced...)
	args = append(args,
		"-map", "0:a", "-map", "[dtstamp]", "-c:a", "pcm_s16le",
		"-ar", "48000", "-ac", "1", "-c:v", "libx264", "-preset",
		state.Runtime.Record_Compression)

	Debug("Compiled recorder arguments: %s", args)
	return args
}
