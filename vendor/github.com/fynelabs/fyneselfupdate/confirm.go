package fyneselfupdate

import (
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"golang.org/x/net/context"
)

// NewUpgradeConfirmCallback creates a GUI based confirmation to use as selfupdate callback
func NewUpgradeConfirmCallback(win fyne.Window) func(string) bool {
	return NewUpgradeConfirmCallbackWithTimeout(win, 0)
}

// NewUpgradeConfirmCallbackWithTimeout creates a GUI based confirmation callback
// with a timeout, after which time the question will be confirmed.
// This can assist in a "default to yes" update where computer may be unattended.
func NewUpgradeConfirmCallbackWithTimeout(win fyne.Window, timeout time.Duration) func(string) bool {
	return func(info string) bool {
		var cancel func()
		var d dialog.Dialog

		var timedout int32

		resp := make(chan bool)
		if timeout > 0 {
			ctx, fn := context.WithTimeout(context.Background(), timeout)
			cancel = fn
			go func() {
				<-ctx.Done()
				atomic.StoreInt32(&timedout, 1)
				d.Hide()
			}()
		}
		d = dialog.NewConfirm("Application Update", info+"\n\nDo you wish to update?\n", func(ok bool) {
			if atomic.LoadInt32(&timedout) == 1 {
				ok = true
			}
			resp <- ok
			if cancel != nil {
				cancel()
			}
		}, win)

		d.Show()
		return <-resp
	}
}

// NewRestartConfirmCallback create a GUI based confirmation to approve restarting the application after being updated
func NewRestartConfirmCallback(win fyne.Window) func() bool {
	return NewRestartConfirmCallbackWithTimeout(win, 0)
}

// NewRestartConfirmCallbackWithTimeout creates a GUI based restarting confirmation callback
// with a timeout, after which time the question will be confirmed to yes.
// This can assist in a "default to yes" update where computer may be unattended.
func NewRestartConfirmCallbackWithTimeout(win fyne.Window, timeout time.Duration) func() bool {
	return func() bool {
		var cancel func()
		var d dialog.Dialog

		var timedout int32

		resp := make(chan bool)
		if timeout > 0 {
			ctx, fn := context.WithTimeout(context.Background(), timeout)
			cancel = fn
			go func() {
				<-ctx.Done()
				atomic.StoreInt32(&timedout, 1)
				d.Hide()
			}()
		}
		d = dialog.NewConfirm("Application Update", "The application was updated.\nDo you wish to restart it?\n", func(ok bool) {
			if atomic.LoadInt32(&timedout) == 1 {
				ok = true
			}
			resp <- ok
			if cancel != nil {
				cancel()
			}
		}, win)

		d.Show()
		return <-resp
	}
}
