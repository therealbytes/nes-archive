package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/fogleman/nes/nes"
)

// Global state variables

//go:embed state/mario.static
var staticData []byte

//go:embed state/mario.dyn
var dynamicData []byte

var runner *nes.Headless

func main() {
	// Expose Go functions to JavaScript
	js.Global().Set("setup", js.FuncOf(setup))
	js.Global().Set("step", js.FuncOf(step))
	js.Global().Set("getState", js.FuncOf(getState))

	<-make(chan bool)
}

// Wrapper functions to handle JavaScript interactions

func setup(this js.Value, args []js.Value) interface{} {
	// staticData = make([]byte, args[0].Length())
	// js.CopyBytesToGo(staticData, args[0])

	// dynamicData = make([]byte, args[1].Length())
	// js.CopyBytesToGo(dynamicData, args[1])

	var err error
	runner, err = nes.NewHeadless(staticData, dynamicData)
	if err != nil {
		panic(err)
	}

	return nil
}

func step(this js.Value, args []js.Value) interface{} {
	startTime := time.Now()

	buttonsData := args[0]
	buttons := make([]uint8, 8)
	js.CopyBytesToGo(buttons, buttonsData)

	endTime := time.Now()
	elapsed := endTime.Sub(startTime)
	fmt.Println("[wasm] Time to copy bytes to Go:", elapsed)

	startTime = time.Now()

	var buttonsBool [8]bool
	for i, button := range buttons {
		buttonsBool[i] = button == 1
	}

	endTime = time.Now()
	elapsed = endTime.Sub(startTime)
	fmt.Println("[wasm] Time to convert bytes to bool:", elapsed)

	startTime = time.Now()

	runner.Console.Controller1.SetButtons(buttonsBool)
	runner.Tick(10)

	endTime = time.Now()
	elapsed = endTime.Sub(startTime)
	fmt.Println("[wasm] Time to tick:", elapsed)

	startTime = time.Now()

	js.CopyBytesToJS(args[1], runner.Console.Buffer().Pix)

	endTime = time.Now()
	elapsed = endTime.Sub(startTime)
	fmt.Println("[wasm] Time to copy bytes to JS:", elapsed)

	return nil
}

func getState(this js.Value, args []js.Value) interface{} {
	staticJSON, _ := json.Marshal(staticData)
	dynamicJSON, _ := json.Marshal(dynamicData)

	return js.ValueOf(map[string]interface{}{
		"staticData":  string(staticJSON),
		"dynamicData": string(dynamicJSON),
	})
}
