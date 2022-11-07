package fyne

import (
	"crypto/ed25519"
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"

	"github.com/fynelabs/fyneselfupdate"
	"github.com/fynelabs/selfupdate"

	"github.com/andydotxyz/fynegameboy/util"
)

type LCD struct {
	Open       func(fyne.URIReadCloser)
	Pause      func()
	Reset      func()
	Resume     func()
	DrawSignal chan bool

	app    fyne.App
	pixels *[160][144][3]uint8
	screen *image.RGBA

	bg      *canvas.Rectangle
	buttons fyne.CanvasObject
	frame   fyne.CanvasObject

	inputStatus *byte
	interrupt   bool
	title       string
	paused      bool
}

func NewDriver() *LCD {
	a := app.NewWithID("xyz.andy.gameboy")
	a.SetIcon(resourceIconPng)
	return &LCD{app: a}
}

func (lcd *LCD) Init(pixels *[160][144][3]uint8, title string) {
	lcd.pixels = pixels
	lcd.title = title
	log.Println("[Display] Initialize Fyne GUI display")
}

func (lcd *LCD) InitStatus(statusPointer *byte) {
	lcd.inputStatus = statusPointer
}

func (lcd *LCD) UpdateInput() bool {
	if lcd.interrupt {
		lcd.interrupt = false

		return true
	}

	return false
}

func (lcd *LCD) NewInput(b []byte) {
}

func (lcd *LCD) draw() image.Image {
	i := 0
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			r, g, b := lcd.pixels[x][y][0], lcd.pixels[x][y][1], lcd.pixels[x][y][2]

			if r == 0xFF && g == 0xFF && b == 0xFF {
				lcd.screen.Pix[i] = 0x9b
				lcd.screen.Pix[i+1] = 0xbc
				lcd.screen.Pix[i+2] = 0x0f
			} else if r == 0xCC && g == 0xCC && b == 0xCC {
				lcd.screen.Pix[i] = 0x8b
				lcd.screen.Pix[i+1] = 0xac
				lcd.screen.Pix[i+2] = 0x0f
			} else if r == 0x77 && g == 0x77 && b == 0x77 {
				lcd.screen.Pix[i] = 0x30
				lcd.screen.Pix[i+1] = 0x62
				lcd.screen.Pix[i+2] = 0x30
			} else {
				lcd.screen.Pix[i] = 0x0f
				lcd.screen.Pix[i+1] = 0x38
				lcd.screen.Pix[i+2] = 0x0f
			}
			lcd.screen.Pix[i+3] = 0xff

			i += 4
		}
	}

	return lcd.screen
}

// Mapping from keys to GB index.
// Reference :https://github.com/Humpheh/goboy/blob/master/pkg/gbio/iopixel/pixels.go
var keyMap = map[fyne.KeyName]byte{
	// A button
	fyne.KeyX: 4,
	// B button
	fyne.KeyZ: 5,
	// SELECT button
	fyne.KeyBackspace: 6,
	// START button
	fyne.KeyReturn: 7,
	// RIGHT button
	fyne.KeyRight: 0,
	// LEFT button
	fyne.KeyLeft: 1,
	// UP button
	fyne.KeyUp: 2,
	// DOWN button
	fyne.KeyDown: 3,
}

func (lcd *LCD) downCode(num uint) {
	statusCopy := *lcd.inputStatus

	statusCopy = util.ClearBit(statusCopy, num)
	lcd.interrupt = true

	*lcd.inputStatus = statusCopy
}

func (lcd *LCD) buttonDown(ev *fyne.KeyEvent) {
	if offset, ok := keyMap[ev.Name]; ok {
		lcd.downCode(uint(offset))
	}
}

func (lcd *LCD) upCode(num uint) {
	statusCopy := *lcd.inputStatus

	statusCopy = util.SetBit(statusCopy, num)
	lcd.interrupt = true

	*lcd.inputStatus = statusCopy
}

func (lcd *LCD) buttonUp(ev *fyne.KeyEvent) {
	if offset, ok := keyMap[ev.Name]; ok {
		lcd.upCode(uint(offset))
	}
}

func (lcd *LCD) MinSize([]fyne.CanvasObject) fyne.Size {
	if fyne.CurrentDevice().IsMobile() {
		return fyne.NewSize(520, 800)
	}

	return fyne.NewSize(520, 400)
}

func (lcd *LCD) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	lcd.bg.Resize(size)

	frameSpacePos := fyne.NewPos(0, 0)
	frameSpaceSize := size
	if fyne.CurrentDevice().IsMobile() {
		buttonSpacePos := fyne.NewPos(0, 0)
		buttonSpaceSize := size
		if fyne.IsHorizontal(fyne.CurrentDevice().Orientation()) {
			oldWidth := frameSpaceSize.Width
			frameSpaceSize.Width *= 0.7
			frameSpacePos.X = (oldWidth - frameSpaceSize.Width) / 2
		} else {
			frameSpaceSize.Height /= 2
			buttonSpaceSize.Height = size.Height - frameSpaceSize.Height
			buttonSpacePos.Y = size.Height - buttonSpaceSize.Height
		}

		lcd.buttons.Move(buttonSpacePos)
		lcd.buttons.Resize(buttonSpaceSize)
	}

	frameSpaceRatio := frameSpaceSize.Width / frameSpaceSize.Height
	frameRatio := float32(1.3)
	frameSize := frameSpaceSize
	framePos := frameSpacePos
	if frameSpaceRatio > frameRatio {
		frameSize = fyne.NewSize(frameSpaceSize.Height*frameRatio, frameSpaceSize.Height)
		framePos = frameSpacePos.Add(fyne.NewPos((frameSpaceSize.Width-frameSize.Width)/2, 0))
	} else if frameSpaceRatio < frameRatio {
		frameSize = fyne.NewSize(frameSpaceSize.Width, frameSpaceSize.Width/frameRatio)
		framePos = frameSpacePos.Add(fyne.NewPos(0, (frameSpaceSize.Height-frameSize.Height)/2))
	}

	lcd.frame.Move(framePos)
	lcd.frame.Resize(frameSize)
}

func (lcd *LCD) Run(drawSignal chan bool, onQuit func()) {
	win := lcd.app.NewWindow(fmt.Sprintf("GameBoy - %s", lcd.title))
	lcd.app.Lifecycle().SetOnExitedForeground(func() {
		if lcd.paused {
			return
		}
		lcd.paused = true
		lcd.Pause()
		d := dialog.NewInformation("Paused", "Tap 'OK' to resume", win)
		d.SetOnClosed(func() {
			lcd.Resume()
			lcd.paused = false
		})
		d.Show()
	})

	lcd.DrawSignal = drawSignal
	lcd.screen = image.NewRGBA(image.Rect(0, 0, 160, 144))
	output := canvas.NewImageFromImage(lcd.screen)
	output.ScaleMode = canvas.ImageScalePixels

	go func() {
		for {
			// drawSignal was sent by the emulator
			<-lcd.DrawSignal

			lcd.draw()
			canvas.Refresh(output)
		}
	}()

	frame := newFrame(output)
	lcd.frame = frame.makeUI()

	lcd.bg = canvas.NewRectangle(&color.Gray{Y: 0xbd})
	objects := []fyne.CanvasObject{lcd.bg, lcd.frame}
	if fyne.CurrentDevice().IsMobile() {
		buttons := newButtons(lcd)
		lcd.buttons = buttons.makeUI()
		objects = append(objects, lcd.buttons)
	}
	content := container.New(lcd, objects...)

	win.SetPadded(false)
	win.SetContent(content)
	if deskCanvas, ok := win.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(lcd.buttonDown)
		deskCanvas.SetOnKeyUp(lcd.buttonUp)
	}
	if !fyne.CurrentDevice().IsMobile() {
		win.Resize(fyne.NewSize(520, 400))
	}

	win.SetOnClosed(func() {
		onQuit()
		lcd.app.Quit()
	})
	win.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File",
		fyne.NewMenuItem("Open...", func() {
			open := dialog.NewFileOpen(func(u fyne.URIReadCloser, err error) {
				if u == nil {
					return
				}
				if err != nil {
					dialog.ShowError(err, win)
					return
				}

				lcd.app.Preferences().SetString("RomURI", u.URI().String())
				lcd.Open(u)
			}, win)
			if !fyne.CurrentDevice().IsMobile() {
				open.SetFilter(storage.NewExtensionFileFilter([]string{".gb"}))
			}
			open.Show()
		}),
		fyne.NewMenuItem("Reset...", func() {
			dialog.ShowConfirm("Reset game", "Are you sure you want to reset?", func(ok bool) {
				if ok {
					lcd.Reset()
				}
			}, win)
		}),
	)))

	selfManage(lcd.app, win)
	win.ShowAndRun()
}


// selfManage turns on automatic update
func selfManage(a fyne.App, w fyne.Window) {
	publicKey := ed25519.PublicKey{173, 172, 119, 135, 110, 164, 228, 191, 220, 253, 163, 146, 202, 189, 212, 148, 155, 64, 68, 156, 181, 255, 111, 120, 36, 254, 212, 232, 210, 15, 155, 101}

	// The public key above match the signature of the below file served by our CDN
	httpSource := selfupdate.NewHTTPSource(nil, "https://geoffrey-test-artefacts.fynelabs.com/self-update/24/24bf6b9b-c379-44b6-9952-3caa0fe81cce/{{.OS}}-{{.Arch}}/{{.Executable}}{{.Ext}}")

	config := fyneselfupdate.NewConfigWithTimeout(a, w, time.Duration(1)*time.Minute,
		httpSource,
		selfupdate.Schedule{FetchOnStart: true, Interval: time.Hour * time.Duration(12)},
		publicKey)

	_, err := selfupdate.Manage(config)
	if err != nil {
		log.Println("Error while setting up update manager: ", err)
		return
	}
}
