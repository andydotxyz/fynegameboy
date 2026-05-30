package main

import (
	"flag"
	"io/ioutil"
	"path/filepath"

	fyneAPI "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/andydotxyz/fynegameboy/fyne"
	"github.com/andydotxyz/fynegameboy/gb"
)

var (
	h bool

	romPath string
	SoundOn bool
	FPS     int
	Debug   bool
)

func init() {
	flag.BoolVar(&h, "h", false, "This help")
	flag.BoolVar(&SoundOn, "m", true, "Turn on sound in GUI mode")
	flag.BoolVar(&Debug, "d", false, "Use Debugger in GUI mode")
	flag.IntVar(&FPS, "f", 60, "Set the `FPS` in GUI mode")
}

func loadRom(romPath string) ([]byte, fyneAPI.URI) {
	data, err := ioutil.ReadFile(romPath)
	if err != nil {
		return nil, nil
	}

	return data, storage.NewFileURI(romPath)
}

func newCore(d *fyne.LCD) *gb.Core {
	core := new(gb.Core)

	core.FPS = FPS
	core.Clock = 4194304
	core.Debug = Debug
	core.DisplayDriver = d
	core.Controller = d
	core.DrawSignal = make(chan bool)
	core.SpeedMultiple = 0
	core.ToggleSound = SoundOn

	return core
}

func startGUI() {
	d := fyne.NewDriver()

	uri, _ := storage.ParseURI(fyneAPI.CurrentApp().Preferences().String("RomURI"))

	var data []byte
	if romPath == "" && uri != nil && uri.String() != "" {
		read, err := storage.Reader(uri)
		if err == nil {
			data, _ = ioutil.ReadAll(read)
		}
	}

	var core *gb.Core
	start := func(resume bool) {
		core = newCore(d)
		d.DrawSignal = core.DrawSignal

		if data != nil && len(data) > 0 {
			core.Init(data, uri)
		} else {
			core.Init(loadRom(romPath))
		}

		if resume {
			core.LoadState()
		}

		go core.Run()
	}
	start(true)

	const volumeControlMemoryLocation = 0xFF26
	// silence the audio channels before tearing the current core down.
	silence := func() {
		core.Sound.Trigger(volumeControlMemoryLocation, 0, core.Memory.MainMemory[0xFF10:0xFF40])
	}
	d.Reset = func() {
		silence()
		core.Exit = true
		start(false)
	}
	d.Open = func(r fyneAPI.URIReadCloser) {
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			fyneAPI.LogError("Unable to load ROM", err)
			return
		}
		_ = r.Close()

		// Persist the outgoing game's state before switching cartridge.
		core.SaveState()
		silence()
		core.Exit = true
		data = bytes
		uri = r.URI()
		start(true)
	}
	d.Save = func() {
		core.SaveState()
	}
	d.ClearState = func() {
		core.DeleteState()
		silence()
		core.Exit = true
		start(false)
	}

	d.Pause = func() {
		core.Pause()
		core.Sound.Trigger(volumeControlMemoryLocation, 0, core.Memory.MainMemory[0xFF10:0xFF40])
	}
	d.Resume = func() {
		data := core.Memory.MainMemory[volumeControlMemoryLocation]
		core.Sound.Trigger(volumeControlMemoryLocation, data, core.Memory.MainMemory[0xFF10:0xFF40])
		core.Resume()
	}

	d.Run(core.DrawSignal, func() {
		core.SaveRAM()
		core.SaveState()
	})
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}

	if len(flag.Args()) == 1 { // probably a ROM parameter
		romPath, _ = filepath.Abs(flag.Arg(0))
	}

	startGUI()
}
