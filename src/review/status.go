package review

import (
	// Standard
	"image/color"

	// 3rd-Party
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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

	status_color := canvas.NewRectangle(color.Transparent)
	status_color.StrokeColor = color.NRGBA{R: 143, G: 176, B: 202, A: 255}
	status_color.StrokeWidth = 3.0

	// 3. Layer the HBox *inside padding* on top of the border Rect
	// NewPadded creates the gap that makes the border Rect visible.
	return container.NewStack(
		status_color,
		container.NewPadded(
			container.NewHBox(
				layout.NewSpacer(),
				status_label,
				layout.NewSpacer())))
}
