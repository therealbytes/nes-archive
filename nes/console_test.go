package nes

import (
	"bytes"
	"testing"
)

func TestConsole(t *testing.T) {
	console1, err := NewConsole("../roms/mario.nes")
	if err != nil {
		t.Fatal(err)
	}
	console1.Reset()

	for i := 0; i < 100_000; i++ {
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

	console2, err := NewHeadlessConsole(static, dynamic)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100_000; i++ {
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
