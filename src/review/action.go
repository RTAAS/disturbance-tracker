// +build !headless

package review

import (
	// DTrack
	"dtrack/ffmpeg"
	"dtrack/log"
	"dtrack/state"

	// Standard
	"fmt"
	"image/png"
	"io"
	"os"
	"strconv"

	// 3rd-Party
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Display an informational message
func Popup(message string) {
	dialog.ShowInformation("OOPS ...", message, get_window(0))
}

// Reset review environment to default
func reset_environment() {
	Current_Status.Set("Step #1: Select a video to open for review.")
	Current_Image.Resource = theme.MoveUpIcon()
	Current_Image.Image = nil
	Current_Image.Refresh()
	Current_Segments.Set([]string{})
	Loaded_Video = make([]VideoSegment, 0)
	Readiness = 0
}

// Reset review environment with custom message and warning icon
func reset_broken_environment(message string) {
	Current_Status.Set(message)
	Current_Image.Resource = theme.WarningIcon()
	Current_Image.Image = nil
	Current_Image.Refresh()
	Current_Segments.Set([]string{})
	Loaded_Video = make([]VideoSegment, 0)
	Readiness = 0
}

// Load selected video into review session
func open_video(uri fyne.URIReadCloser, err error) {
	// User cancelled or no file selected
	if uri == nil || err != nil {
		return
	}
	defer uri.Close()
	mkvpath := uri.URI()
	mkvData := make([]VideoSegment, 0)
	stdReader, stdWriter := io.Pipe()
	var segment_id uint = 0
	var idList []string

	// Create temporary temporary directory for mkv extraction
	extractDir, err := os.MkdirTemp("", "dtrack_*")
	if err != nil {
		reset_broken_environment("Failed to load video!\nError: " + err.Error())
		return
	}
	if !state.Runtime.Workspace_Keep_Temp {
		defer os.RemoveAll(extractDir)
		log.Trace("Unpacking to temporary directory: %s", extractDir)
	} else {
		log.Warn("%s will not be removed after extraction", extractDir)
	}

	// Extract images while reading wav stream
	args := ffmpeg.Extract_Arguments(mkvpath.Path(), extractDir)
	go ffmpeg.ReadStdin(args, stdWriter, true)
	for {
		// Allocate a buffer for the audio segment
		segment_data := make([]byte, ffmpeg.BytesPerSecond)

		// Block until segment_data is full
		_, err := io.ReadFull(stdReader, segment_data)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			log.Trace("Encountered end of wav stream")
			break
		}
		if err != nil {
			reset_broken_environment("Unhandled error: " + err.Error())
		}

		// Add new segment to list
		log.Trace("New segment read: %d", segment_id)
		idList = append(idList, strconv.FormatUint(uint64(segment_id), 10))
		mkvData = append(mkvData, VideoSegment{
			count: segment_id,
			data:  segment_data})
		segment_id++
	}

	// Load images and wav clips into movie data
	for i := 0; i < len(mkvData); i++ {
		// Image
		imagePath := fmt.Sprintf("%s/%d.png", extractDir, i)
		log.Trace("Loading image: %s", imagePath)
		fh, err := os.Open(imagePath)
		if err != nil {
			log.Die("Open Error: %s", err)
		}
		mkvData[i].image, err = png.Decode(fh)
		if err != nil {
			log.Die("Decode Error: %s", err)
		}
		fh.Close()

		// Audio
		audioPath := fmt.Sprintf("%s/%d.wav", extractDir, i)
		log.Trace("Loading audio: %s", audioPath)
		content, err := os.ReadFile(audioPath)
		if err != nil {
			log.Die("Open Error: %s", err)
		}
		mkvData[i].audio = content
		fh.Close()
	}

	// Merge loaded video into review session
	Loaded_Video = mkvData
	Current_Filename = mkvpath.Name()
	// Exclude the last, because it has no trailing audio to consume
	Current_Segments.Set(idList[:len(idList)-1])
	Readiness = 1

	// Display next step
	Current_Status.Set("Step #2: Select a recording clip to review.")
	Current_Image.Resource = theme.NavigateBackIcon()
	Current_Image.Image = nil
	Current_Image.Refresh()
}

// Load one clip into current review session
func load_clip(id widget.ListItemID) {
	if Readiness < 1 {
		Current_Status.Set("ERROR: No video loaded ...")
		return
	}
	segments, err := Current_Segments.Get()
	if err != nil {
		log.Warn("Unexpected index selected!")
		return
	}
	log.Debug("Loading clip: %s", segments[id])

	if start, err := strconv.Atoi(segments[id]); err != nil {
		Current_Status.Set("Error loading clip: " + err.Error())
	} else {
		Current_Frame = start
		Current_Image.Image = Loaded_Video[start].image
		Current_Image.Resource = nil
		Current_Image.Refresh()
		// Use goroutine to allow immediate refresh
		go play_selected()
		Current_Status.Set("Step #3: Carefully listen to this audio clip.")
		Readiness = 2
	}
}

// Re-play_selected() segment, then update status with next step
func replay_segment() {
	play_selected()
	Current_Status.Set("Step #4: Save clip using appropriate tag.")
	Readiness = 3
}

// Play wav file from current and next frame (2 frames -> 1 segment)
func play_selected() {
	if Readiness < 2 {
		Popup("No clip selected.")
		return
	}
	// TODO: This creates lag between clips; would be nice to concatenate
	ffmpeg.PlayData(Loaded_Video[Current_Frame].audio)
	ffmpeg.PlayData(Loaded_Video[Current_Frame+1].audio)
}

// Copy audio segment (raw data) to tag directory
func tag_clip(tag string) {
	switch {
	case Readiness < 2:
		Popup("No clip selected.")
		return
	case Readiness == 2:
		Popup("Replay audio at least once.")
		return
	case Readiness >= 4:
		Popup("Already tagged.")
		return
	}
	tagDir := fmt.Sprintf("%s/tags/%s/", state.Runtime.Workspace, tag)
	tagFile := fmt.Sprintf("%s/%s:%d.dat", tagDir, Current_Filename, Current_Frame)

	// Ensure output directory exists
	if err := os.MkdirAll(tagDir, 0755); err != nil {
		Current_Status.Set("Error creating directory:\n" + err.Error())
		return
	}

	// Create file from audio data
	data := append(
		Loaded_Video[Current_Frame].data,
		Loaded_Video[Current_Frame+1].data...)
	if err := os.WriteFile(tagFile, data, 0644); err != nil {
		Current_Status.Set("Error writing file:\n" + err.Error())
		return
	}

	// Notify of completion
	Current_Status.Set("Audio segment tagged!")
	log.Debug("Audio segment saved as %s", tagFile)
	Readiness = 4
}

func select_model_and_train() {
	// TODO: Consider making a pretty display that shows podman stdout/err
}
