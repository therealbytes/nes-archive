package nes

import (
	"bytes"
	"testing"
)

func TestSerialization(t *testing.T) {
	console1, err := NewConsole("../roms/mario.nes")
	if err != nil {
		t.Fatal(err)
	}
	console1.Reset()

	for i := 0; i < 1000; i++ {
		console1.Step()
	}

	static, err := console1.SerializeStatic()
	if err != nil {
		t.Fatal(err)
	}

	dynamic, err := console1.SerializeDynamic()
	if err != nil {
		t.Fatal(err)
	}

	cartridge := &Cartridge{}
	ram := make([]byte, 2048)
	controller1 := NewController()
	controller2 := NewController()
	console2 := Console{nil, nil, nil, cartridge, controller1, controller2, nil, ram}

	err = console2.DeserializeStatic(static)
	if err != nil {
		t.Fatal(err)
	}

	mapper, err := NewMapper(&console2)
	if err != nil {
		t.Fatal(err)
	}
	console2.Mapper = mapper
	console2.CPU = NewCPU(&console2)
	console2.APU = NewAPU(&console2)
	console2.PPU = NewPPU(&console2)

	err = console2.DeserializeDynamic(dynamic)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		console1.Step()
		console2.Step()
	}

	if console1.CPU.PC != console2.CPU.PC {
		t.Fatal("PCs are not equal")
	}

	if !bytes.Equal(console1.RAM, console2.RAM) {
		t.Fatal("RAM are not equal")
	}

	if !bytes.Equal(console1.Buffer().Pix, console2.Buffer().Pix) {
		t.Fatal("buffers are not equal")
	}
}
