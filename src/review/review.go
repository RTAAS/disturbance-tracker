package review

import (
	// 3rd-Party
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/theme"
)

// List of audio segments available in loaded recording
var Current_Segments binding.StringList = binding.NewStringList()

// Image currently displayed in preview pane
var Current_Image *canvas.Image = canvas.NewImageFromImage(nil)

// Main structure of review window
func review_window() fyne.CanvasObject {
	top, middle, bottom := menu_bar(), segment_viewer(), status_bar()
	return container.New(
		layout.NewBorderLayout(top, bottom, nil, nil),
		top, middle, bottom)
}

// Return the primary review container
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
	//segment_list.OnSelected = func(id widget.ListItemID) {
	//	Current_Status.Set("Selected clip: " + Current_Segments[id])
	//}

	train_button := widget.NewButton(
		"[ Step #5 ]\nBegin Training",
		 select_model_and_train)

	// Assemble left-hand vertical stack
	left_side := container.New(
		layout.NewBorderLayout(segment_label, train_button, nil, nil),
		segment_label, segment_list, train_button)

	// Right: Image showing first frame of video segment
	Current_Image.FillMode = canvas.ImageFillOriginal
	Current_Image.Resource = theme.NavigateBackIcon()

	// Assemble the actual workspace area
	body := container.NewHSplit(
		left_side,
		Current_Image)
	body.SetOffset(0.22)
	return body
}

// Load user-selected video
func open_video(uri fyne.URIReadCloser, err error) {
	// User cancelled or no file selected
	if uri == nil || err != nil {
		return
	}
	defer uri.Close()
	filepath := uri.URI().Path()

	// Mock update
	Current_Status.Set("Selected: " + filepath)
	//Current_Segments.Set([]string{"1", "2", "3", "4", "5", "6", "7"})
	//Current_Image.Resource = theme.FyneLogo()
}

// Replay currently selected audio segment
func replay_segment() {
}

// Copy (1-second) audio segment (.wav) to tag directory
func tag_clip() {
	// TODO
}

func select_model_and_train() {
}
