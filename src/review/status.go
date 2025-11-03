package review

import (
	// 3rd-Party
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Holds text stored in the status bar
var Current_Status binding.String = binding.NewString()

// Label that is modified via Update_Status(text)
func status_bar() fyne.CanvasObject {
	status_label := widget.NewLabelWithData(Current_Status)
	status_label.TextStyle.Bold = true

	return container.NewHBox(
		layout.NewSpacer(),
		status_label,
		layout.NewSpacer(),
	)
}
