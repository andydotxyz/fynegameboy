package gb

import (
	"bytes"
	"encoding/gob"
	"testing"
)

// newTestCore builds a minimal core backed by an MBC1 with empty cartridge RAM.
func newTestCore(t *testing.T) *Core {
	t.Helper()
	core := &Core{}
	core.Cartridge.MBC = &MBC1{
		rom:            make([]byte, 0x8000),
		CurrentROMBank: 1,
		RAMBank:        make([]byte, 0x8000),
	}
	return core
}

func TestSaveStateRoundTrip(t *testing.T) {
	src := newTestCore(t)
	// Populate state across all the regions a save-state must capture.
	src.Memory.MainMemory[0xC000] = 0x42 // work RAM
	src.Memory.MainMemory[0xFF80] = 0x99 // high RAM
	src.CPU.Registers.PC = 0x1234
	src.CPU.Registers.SP = 0xFFF0
	src.CPU.Registers.A = 0xAB
	src.CPU.Flags.Zero = true
	src.CPU.Halt = true
	src.Timer.ScanlineCounter = 321
	src.JoypadStatus = 0xCD
	src.SpeedMultiple = 1
	mbc := src.Cartridge.MBC.(*MBC1)
	mbc.CurrentROMBank = 5
	mbc.CurrentRAMBank = 2
	mbc.EnableRAM = true
	mbc.RAMBank[0x10] = 0x7E // cartridge external RAM

	data, err := src.marshalState()
	if err != nil {
		t.Fatalf("marshalState failed: %v", err)
	}

	// A fresh core that has only just powered on (defaults differ from src).
	dst := newTestCore(t)
	if !dst.applyState(data) {
		t.Fatal("applyState returned false for a valid snapshot")
	}

	if dst.Memory.MainMemory[0xC000] != 0x42 || dst.Memory.MainMemory[0xFF80] != 0x99 {
		t.Errorf("MainMemory not restored: got %#x / %#x", dst.Memory.MainMemory[0xC000], dst.Memory.MainMemory[0xFF80])
	}
	if dst.CPU.Registers.PC != 0x1234 || dst.CPU.Registers.SP != 0xFFF0 || dst.CPU.Registers.A != 0xAB {
		t.Errorf("CPU registers not restored: %+v", dst.CPU.Registers)
	}
	if !dst.CPU.Flags.Zero || !dst.CPU.Halt {
		t.Errorf("CPU flags/halt not restored: %+v halt=%v", dst.CPU.Flags, dst.CPU.Halt)
	}
	if dst.Timer.ScanlineCounter != 321 || dst.JoypadStatus != 0xCD || dst.SpeedMultiple != 1 {
		t.Errorf("misc state not restored: timer=%d joypad=%#x speed=%d", dst.Timer.ScanlineCounter, dst.JoypadStatus, dst.SpeedMultiple)
	}
	dmbc := dst.Cartridge.MBC.(*MBC1)
	if dmbc.CurrentROMBank != 5 || dmbc.CurrentRAMBank != 2 || !dmbc.EnableRAM || dmbc.RAMBank[0x10] != 0x7E {
		t.Errorf("MBC state not restored: rom=%d ram=%d enable=%v cell=%#x",
			dmbc.CurrentROMBank, dmbc.CurrentRAMBank, dmbc.EnableRAM, dmbc.RAMBank[0x10])
	}
}

func TestApplyStateRejectsInvalid(t *testing.T) {
	core := newTestCore(t)
	core.Memory.MainMemory[0xC000] = 0x11

	// Empty and garbage payloads must be rejected without mutating the core.
	if core.applyState(nil) {
		t.Error("applyState accepted nil data")
	}
	if core.applyState([]byte{0x01, 0x02, 0x03}) {
		t.Error("applyState accepted garbage data")
	}

	// A snapshot from an incompatible version must be ignored.
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&CoreState{Version: saveStateVersion + 1}); err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if core.applyState(buf.Bytes()) {
		t.Error("applyState accepted an incompatible version")
	}

	if core.Memory.MainMemory[0xC000] != 0x11 {
		t.Error("applyState mutated core while rejecting invalid data")
	}
}
