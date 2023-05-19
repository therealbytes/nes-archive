package main

import (
	"encoding/json"
	"image"
	"image/color"
	"syscall/js"
)

// Global state variables
var staticData []byte
var dynamicData []byte

func main() {
	// Expose Go functions to JavaScript
	js.Global().Set("setup", js.FuncOf(setup))
	js.Global().Set("step", js.FuncOf(step))
	js.Global().Set("getState", js.FuncOf(getState))

	<-make(chan bool)
}

// Wrapper functions to handle JavaScript interactions

func setup(this js.Value, args []js.Value) interface{} {
	staticData = make([]byte, args[0].Length())
	js.CopyBytesToGo(staticData, args[0])

	dynamicData = make([]byte, args[1].Length())
	js.CopyBytesToGo(dynamicData, args[1])

	return nil
}

func step(this js.Value, args []js.Value) interface{} {
	action := args[0]
	actionBytes := make([]byte, action.Length())
	js.CopyBytesToGo(actionBytes, action)

	// Perform some image processing or calculations here...

	// Return the image data
	width := 256
	height := 240
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{0, uint8(x % 256), uint8(y % 256), 255})
		}
	}

	js.CopyBytesToJS(args[1], img.Pix)

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
