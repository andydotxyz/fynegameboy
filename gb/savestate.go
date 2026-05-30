package gb

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

/*
Save-states capture the entire running machine so that reopening a ROM
resumes execution exactly where it left off. This is distinct from the
battery .sav file (see SaveRAM), which only persists the cartridge's
external RAM. Save-states are written to a ".st" file alongside the ROM.

Bumping saveStateVersion invalidates older incompatible files, which are
then ignored (the game simply boots fresh) rather than crashing.
*/
const saveStateVersion = 1

/*
CoreState holds every piece of volatile state needed to resume emulation.
The Game Boy's working RAM (VRAM, WRAM, OAM, HRAM and I/O registers) all
lives in MainMemory, but the CPU registers, timer counters and the MBC's
bank selection / cartridge RAM live outside it, so they are captured here
too. Only exported fields are serialised by encoding/gob.
*/
type CoreState struct {
	Version        int
	MainMemory     []byte
	CPU            CPU
	Timer          Timer
	JoypadStatus   byte
	SerialByte     byte
	InterruptCount int
	SpeedMultiple  int
	MBC            MBCState
}

/*
SaveState writes a full snapshot of the machine to the ROM's ".st" file.
It is a no-op when no state URI is known (for example the bundled ROM).
*/
func (core *Core) SaveState() {
	if core.StateURI == nil {
		return
	}

	data, err := core.marshalState()
	if err != nil {
		log.Println("Failed to encode save state", err)
		return
	}

	if err := writeURIFile(core.StateURI, data); err != nil {
		log.Println("Failed to write save state", err)
		return
	}
	log.Printf("[Core] Save state written (%d bytes) to %s\n", len(data), core.StateURI)
}

/*
marshalState serialises the current machine state to bytes. Split out from
SaveState so the snapshot logic can be exercised without file I/O.
*/
func (core *Core) marshalState() ([]byte, error) {
	state := CoreState{
		Version:        saveStateVersion,
		MainMemory:     append([]byte(nil), core.Memory.MainMemory[:]...),
		CPU:            core.CPU,
		Timer:          core.Timer,
		JoypadStatus:   core.JoypadStatus,
		SerialByte:     core.SerialByte,
		InterruptCount: core.InterruptCount,
		SpeedMultiple:  core.SpeedMultiple,
		MBC:            core.Cartridge.MBC.SaveState(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&state); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/*
applyState decodes a snapshot produced by marshalState and copies it into
the running machine. Returns false (without mutating the core) when the
data is unreadable or from an incompatible version.
*/
func (core *Core) applyState(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	var state CoreState
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&state); err != nil {
		log.Println("Failed to decode save state", err)
		return false
	}
	if state.Version != saveStateVersion {
		log.Printf("[Core] Ignoring incompatible save state (version %d)\n", state.Version)
		return false
	}

	if len(state.MainMemory) == len(core.Memory.MainMemory) {
		copy(core.Memory.MainMemory[:], state.MainMemory)
	}
	core.CPU = state.CPU
	core.Timer = state.Timer
	core.JoypadStatus = state.JoypadStatus
	core.SerialByte = state.SerialByte
	core.InterruptCount = state.InterruptCount
	core.SpeedMultiple = state.SpeedMultiple
	core.Cartridge.MBC.LoadState(state.MBC)
	return true
}

/*
LoadState restores a snapshot previously written by SaveState, resuming
the machine exactly where it was. If no state file exists, or it is
unreadable or from an incompatible version, the call is a no-op and the
game continues from its freshly initialised (power-on) state.
*/
func (core *Core) LoadState() {
	if core.StateURI == nil {
		return
	}

	read, err := storage.Reader(core.StateURI)
	if err != nil {
		// No save state yet - nothing to resume.
		return
	}
	defer read.Close()

	data, err := ioutil.ReadAll(read)
	if err != nil {
		return
	}

	if core.applyState(data) {
		log.Println("[Core] Save state loaded")
	}
}

/*
DeleteState removes the ROM's save-state file, if present. Combined with a
fresh re-initialisation (see the Reset flow), this discards the captured
machine state so the game boots from power-on next time.
*/
func (core *Core) DeleteState() {
	if core.StateURI == nil {
		return
	}

	exists, err := storage.Exists(core.StateURI)
	if err != nil || !exists {
		return
	}
	if err := storage.Delete(core.StateURI); err != nil {
		log.Println("Failed to delete save state", err)
	}
}

func writeURIFile(uri fyne.URI, data []byte) error {
	write, err := storage.SaveFileToURI(uri)
	if err != nil {
		return err
	}
	defer write.Close()

	_, err = write.Write(data)
	return err
}
