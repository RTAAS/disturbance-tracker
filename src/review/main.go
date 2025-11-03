//##
// DTrack Package: GUI Review
//
// Splits audio (+video optional) files into 2-second clips and provides a
// GUI tool that moves tagged audio clips into special directories, for traning.
//##
package review

import (
	// 3rd-Party
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// Primary post-bootstrap entry point
func Start() {
	application := app.NewWithID("DTrack")
	root_window := application.NewWindow("DTrack Review")
	root_window.Resize(fyne.NewSize(1024, 768))
	root_window.SetContent(review_window())
	Current_Status.Set("Select a video to begin review ...")
	root_window.ShowAndRun()
}

// Returns the base object for a root window
func get_window(root int) fyne.Window {
	return fyne.CurrentApp().Driver().AllWindows()[root]
}
