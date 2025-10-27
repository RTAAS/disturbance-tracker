// +build !headless

// ##
// DTrack Package: GUI Review
//
// Splits audio (+video optional) files into 2-second clips and provides a
// GUI tool that moves tagged audio clips into special directories, for traning.
// ##
package review

import (
	// Standard
	"image"

	// 3rd-Party
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Most recently completed step
var Readiness int = 0

// List of audio segments available in loaded recording
var Current_Segments binding.StringList = binding.NewStringList()

// Image currently displayed in preview pane
var Current_Image *canvas.Image = canvas.NewImageFromImage(nil)

// Currently selected frame
var Current_Frame int

// Use existing filename in output slices
var Current_Filename string

// Collection of all audio and clips from an mkv file
var Loaded_Video []VideoSegment

// Single segment of sliced mkv file
type VideoSegment struct {
	count uint        // Copy of index value
	data  []byte      // Raw audio data, for machine learning
	audio []byte      // One-second wav clip
	image image.Image // Image from one video frame
}

// Primary post-bootstrap entry point
func Launch() {
	application := app.NewWithID("DTrack")
	root_window := application.NewWindow("DTrack Review")
	root_window.Resize(fyne.NewSize(1024, 768))
	root_window.SetContent(review_window())
	reset_environment()
	root_window.ShowAndRun()
}

// Returns the base object for a root window
func get_window(root int) fyne.Window {
	return fyne.CurrentApp().Driver().AllWindows()[root]
}

// Main structure of review window
func review_window() fyne.CanvasObject {
	top, middle, bottom := menu_bar(), status_bar(), segment_viewer()
	return container.New(
		layout.NewBorderLayout(top, nil, nil, nil),
		top,
		container.New(
			layout.NewBorderLayout(middle, nil, nil, nil),
			middle,
			bottom))
}

// Main review container (middle row)
func segment_viewer() fyne.CanvasObject {
	// List of (1-second) video segments
	segment_text := widget.NewLabel("[ Step #2 ]\nSelect Clip:")
	segment_text.TextStyle.Bold = true
	segment_label := container.NewCenter(segment_text)
	segment_list := widget.NewListWithData(
		Current_Segments,
		func() fyne.CanvasObject {
			// Object created for each list item
			return widget.NewLabel("")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			// Bind label text to item value
			o.(*widget.Label).Bind(i.(binding.String))
		})
	// Event: Clicked segment name from list
	segment_list.OnSelected = load_clip

	train_button := widget.NewButton(
		"[ Step #5 ]\nBegin Training",
		select_model_and_train)

	// Assemble left-hand vertical stack
	left_side := container.New(
		layout.NewBorderLayout(segment_label, train_button, nil, nil),
		segment_label, segment_list, train_button)

	// Right: Image showing first frame of video segment
	Current_Image.FillMode = canvas.ImageFillOriginal

	// Assemble the actual workspace area
	body := container.NewHSplit(
		left_side,
		Current_Image)
	body.SetOffset(0.22)
	return body
}
