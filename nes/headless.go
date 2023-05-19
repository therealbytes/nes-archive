package nes

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io/ioutil"
)

// 1789773 Hz / 100_000 = 17.89773 Hz
var ConsoleStepsPerTick = 100_000

type Headless struct {
	Console Console
}

type Action struct {
	Button int
	Press  bool
	Wait   int
}

func NewHeadless(static []byte, dynamic []byte) (*Headless, error) {
	staticDecoder := gob.NewDecoder(bytes.NewBuffer(static))
	dynamicDecoder := gob.NewDecoder(bytes.NewBuffer(dynamic))

	cartridge := &Cartridge{}
	ram := make([]byte, 2048)
	controller1 := NewController()
	controller2 := NewController()
	console := Console{nil, nil, nil, cartridge, controller1, controller2, nil, ram}

	if err := console.LoadStatic(staticDecoder); err != nil {
		return nil, err
	}

	mapper, err := NewMapper(&console)
	if err != nil {
		return nil, err
	}
	console.Mapper = mapper
	console.CPU = NewCPU(&console)
	console.APU = NewAPU(&console)
	console.PPU = NewPPU(&console)

	if err := console.LoadDynamic(dynamicDecoder); err != nil {
		return nil, err
	}

	return &Headless{console}, nil
}

func (headless *Headless) Run(activity []Action) {
	console := headless.console
	for _, action := range activity {
		if action.Button < 8 {
			console.Controller1.buttons[action.Button] = action.Press
		}
		if action.Wait > 0 {
			headless.Tick(action.Wait)
		}
	}
}

// One tick corresponds to a fixed number of CPU Steps (not cycles!).
func (headless *Headless) Tick(ticks int) {
	console := headless.console
	steps := ConsoleStepsPerTick * ticks
	for step := 0; step < steps; step++ {
		console.Step()
	}
}

func Compress(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return ioutil.ReadAll(gz)
}
