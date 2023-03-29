package fyneselfupdate

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// NewExitCallback properly exit a Fyne App. If a window is specified, it will report
// error using dialog.ShowError if the upgrade failed to Restart. If no window is
// specified, it will display error on the command line.
func NewExitCallback(app fyne.App, win fyne.Window) func(err error) {
	if win != nil {
		return func(err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			app.Quit()
		}
	}
	return func(err error) {
		if err != nil {
			log.Printf("Unable to exit to a new version: %v\n", err)
			return
		}
		app.Quit()
	}
}
