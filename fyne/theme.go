//go:generate fyne bundle -package fyne -o bundled.go ../Icon.png

package fyne

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type buttons struct {
	up, down, left, right, middle fyne.CanvasObject
	start, sel, a, b              fyne.CanvasObject
}

func newButtons(lcd *LCD) *buttons {
	up := newGameButton(lcd, 2, canvas.NewRectangle(color.Black))
	down := newGameButton(lcd, 3, canvas.NewRectangle(color.Black))
	left := newGameButton(lcd, 1, canvas.NewRectangle(color.Black))
	right := newGameButton(lcd, 0, canvas.NewRectangle(color.Black))
	mid := canvas.NewCircle(color.Gray{Y: 0x16})
	mid.StrokeColor = color.Black
	mid.StrokeWidth = 16

	a := newGameButton(lcd, 4, canvas.NewCircle(color.NRGBA{R: 0xc9, G: 0x20, B: 0x90, A: 0xFF}))
	b := newGameButton(lcd, 5, canvas.NewCircle(color.NRGBA{R: 0xc9, G: 0x20, B: 0x90, A: 0xFF}))
	start := newGameButton(lcd, 7, startShape())
	sel := newGameButton(lcd, 6, startShape())

	return &buttons{up: up, down: down, left: left, right: right, middle: mid, a: a, b: b, start: start, sel: sel}
}

func (b *buttons) Layout(_ []fyne.CanvasObject, s fyne.Size) {
	scale := float32(math.Min(float64(s.Width/130.0), float64(s.Height/100.0)))

	abSize := fyne.NewSize(17.5*scale, 17.5*scale)
	startSize := fyne.NewSize(21*scale, 5*scale)
	dSize := fyne.NewSize(12.5*scale, 12.5*scale)

	b.a.Resize(abSize)
	b.b.Resize(abSize)
	b.start.Resize(startSize)
	b.sel.Resize(startSize)

	b.up.Resize(dSize)
	b.down.Resize(dSize)
	b.left.Resize(dSize)
	b.right.Resize(dSize)
	b.middle.Resize(fyne.NewSize(16*scale, 16*scale))

	dPadTop, dPadLeft := float32(26.25), float32(4.5)
	if fyne.IsHorizontal(fyne.CurrentDevice().Orientation()) {
		xPad := (s.Width/scale - 200) / 4
		xRightPad := s.Width/scale - xPad
		dPadLeft = xPad + 0.5

		b.a.Move(fyne.NewPos((xRightPad-18)*scale, 19*scale))
		b.b.Move(fyne.NewPos((xRightPad-35.25)*scale, 34.25*scale))

		b.start.Move(fyne.NewPos((xRightPad-21.25)*scale, 70*scale))
		b.sel.Move(fyne.NewPos((xRightPad-37.75)*scale, 80*scale))
	} else {
		xOff := (s.Width/scale - 130) / 2
		yOff := (s.Height/scale - 100) / 2
		b.a.Move(fyne.NewPos((xOff+106.25)*scale, (yOff+29)*scale))
		b.b.Move(fyne.NewPos((xOff+82)*scale, (yOff+41.25)*scale))

		b.start.Move(fyne.NewPos((xOff+57.5)*scale, (yOff+81.25)*scale))
		b.sel.Move(fyne.NewPos((xOff+33.75)*scale, (yOff+81.25)*scale))

		dPadLeft += xOff
		dPadTop += yOff
	}

	b.up.Move(fyne.NewPos(float32(dPadLeft+12.5)*scale, float32(dPadTop)*scale))
	b.down.Move(fyne.NewPos(float32(dPadLeft+12.5)*scale, float32(dPadTop+25)*scale))
	b.left.Move(fyne.NewPos(float32(dPadLeft)*scale, float32(dPadTop+12.5)*scale))
	b.right.Move(fyne.NewPos(float32(dPadLeft+25)*scale, float32(dPadTop+12.5)*scale))
	b.middle.Move(fyne.NewPos(float32(dPadLeft+10.75)*scale, float32(dPadTop+10.75)*scale))
}

func (b *buttons) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(70, 100)
}

func (b *buttons) makeUI() fyne.CanvasObject {
	return container.New(b, b.up, b.down, b.left, b.right, b.middle, b.a, b.b, b.start, b.sel)
}

type frame struct {
	output *canvas.Image

	power, tl, tr, bl, br *canvas.Circle
	b1, b2, b3, b4        *canvas.Rectangle
}

func newFrame(i *canvas.Image) *frame {
	return &frame{output: i}
}

func (f *frame) Layout(_ []fyne.CanvasObject, s fyne.Size) {
	scale := s.Width / 130.0

	f.output.Resize(fyne.NewSize(80*scale, 72*scale))
	f.output.Move(fyne.NewPos(25*scale, 14*scale))

	f.power.Resize(fyne.NewSize(4*scale, 4*scale))
	f.power.Move(fyne.NewPos(9.25*scale, 36.5*scale))

	f.tl.Resize(fyne.NewSize(8*scale, 8*scale))
	f.tl.Move(fyne.NewPos(2*scale, 2*scale))
	f.tr.Resize(fyne.NewSize(8*scale, 8*scale))
	f.tr.Move(fyne.NewPos(120*scale, 2*scale))
	f.bl.Resize(fyne.NewSize(8*scale, 8*scale))
	f.bl.Move(fyne.NewPos(2*scale, 90*scale))
	f.br.Resize(fyne.NewSize(40*scale, 40*scale))
	f.br.Move(fyne.NewPos(88*scale, 58*scale))

	f.b1.Move(fyne.NewPos(6*scale, 2*scale))
	f.b1.Resize(fyne.NewSize(118*scale, 8*scale))
	f.b2.Move(fyne.NewPos(2*scale, 6*scale))
	f.b2.Resize(fyne.NewSize(126*scale, 72*scale))
	f.b3.Move(fyne.NewPos(2*scale, 74*scale))
	f.b3.Resize(fyne.NewSize(106*scale, 20*scale))
	f.b4.Move(fyne.NewPos(6*scale, 90*scale))
	f.b4.Resize(fyne.NewSize(102*scale, 8*scale))
}

func (f *frame) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(130, 100)
}

func (f *frame) makeUI() fyne.CanvasObject {
	f.power = canvas.NewCircle(color.NRGBA{R: 0xd9, G: 0x10, B: 0x10, A: 0xff})

	frameCol := color.NRGBA{R: 0x64, G: 0x66, B: 0x76, A: 0xff}
	f.tl = canvas.NewCircle(frameCol)
	f.tr = canvas.NewCircle(frameCol)
	f.bl = canvas.NewCircle(frameCol)
	f.br = canvas.NewCircle(frameCol)
	f.b1 = canvas.NewRectangle(frameCol)
	f.b2 = canvas.NewRectangle(frameCol)
	f.b3 = canvas.NewRectangle(frameCol)
	f.b4 = canvas.NewRectangle(frameCol)
	return container.New(f, f.tl, f.tr, f.bl, f.br, f.b1, f.b2, f.b3, f.b4, f.output, f.power)
}

type curvedBox struct {
	left, right *canvas.Circle
	mid         *canvas.Rectangle
}

func (c *curvedBox) Layout(_ []fyne.CanvasObject, s fyne.Size) {
	h := s.Height
	c.left.Resize(fyne.NewSize(h, h))
	c.right.Resize(fyne.NewSize(h, h))
	c.mid.Resize(fyne.NewSize(s.Width-h, h))

	c.right.Move(fyne.NewPos(s.Width-h, 0))
	c.mid.Move(fyne.NewPos(h/2, 0))
}

func (c *curvedBox) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(5, 5)
}

func startShape() fyne.CanvasObject {
	btnCol := color.NRGBA{R: 0x66, G: 0x66, B: 0x66, A: 0xFF}
	left := canvas.NewCircle(btnCol)
	right := canvas.NewCircle(btnCol)
	mid := canvas.NewRectangle(btnCol)
	return container.New(&curvedBox{left, right, mid}, left, right, mid)
}
