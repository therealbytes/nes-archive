package nes

import (
	"bytes"
	"testing"
	"time"
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

	console2, err := NewHeadlessConsole(static, dynamic, false)
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

func BenchmarkConsole(b *testing.B) {
	ogConsole, err := NewConsole("../roms/mario.nes")
	if err != nil {
		b.Fatal(err)
	}
	ogConsole.Reset()

	static, err := ogConsole.SerializeStatic()
	if err != nil {
		b.Fatal(err)
	}
	dynamic, err := ogConsole.SerializeDynamic()
	if err != nil {
		b.Fatal(err)
	}

	noAPUConsole, err := NewHeadlessConsole(static, dynamic, true)
	if err != nil {
		b.Fatal(err)
	}

	noPixConsole, err := NewHeadlessConsole(static, dynamic, false)
	if err != nil {
		b.Fatal(err)
	}

	bms := []struct {
		name    string
		console *Console
	}{
		{"APU && PIX", ogConsole},
		{"!APU && PIX", noAPUConsole},
		{"!APU && !PIX", noPixConsole},
	}

	for _, bm := range bms {
		b.Run(bm.name, func(b *testing.B) {
			startTime := time.Now()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.console.StepSeconds(100 * time.Millisecond.Seconds())
			}
			b.StopTimer()

			durationNs := time.Since(startTime).Nanoseconds()
			consoleNs := int64(b.N) * 100 * time.Millisecond.Nanoseconds()

			b.ReportMetric(float64(consoleNs)/float64(durationNs), "x")
		})
	}
}
