package review

import (
	// DTrack
	"dtrack/state"

	// 3rd-Party
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// Row of buttons that function as a main menu
func menu_bar() fyne.CanvasObject {
	// Menu length: 1+1+1+1+X = 4+X
	menu_buttons := make(
		[]fyne.CanvasObject, 0,
		4+len(state.Runtime.Record_Inspect_Models))

	// Select Video [+1]
	menu_buttons = append(menu_buttons,
		widget.NewButton("[ Step #1 ]\nSelect Video", select_video))
	// Replay Audio [+1]
	menu_buttons = append(menu_buttons,
		widget.NewButton("[ Step #3 ]\nReplay Audio", replay_segment))
	// Save Label   [+1]
	save_label := widget.NewLabel("[ Step #4 ]\nSave Clip -->")
	save_label.TextStyle.Bold = true
	menu_buttons = append(menu_buttons,
		container.NewCenter(save_label))
	// Tag no-match [+1]
	menu_buttons = append(menu_buttons,
		widget.NewButton("[ Save as ]\nNo Match", func() {
			tag_clip("empty")
		}))
	// Models [+X]
	for _, model := range state.Runtime.Record_Inspect_Models {
		menu_buttons = append(menu_buttons,
			widget.NewButton("[ Save as ]\n"+model, func() {
				tag_clip(model)
			}))
	}

	// Return buttons inside full-width container
	return container.NewGridWithColumns(len(menu_buttons), menu_buttons...)
}

// Prompt to select MKV from workspace recordings
func select_video() {
	cwd, _ := storage.ListerForURI(storage.NewFileURI(
		state.Runtime.Workspace + "/recordings"))
	// Return selected video to open_video()
	open := dialog.NewFileOpen(open_video, get_window(0))
	open.SetLocation(cwd)
	open.Show()
}
