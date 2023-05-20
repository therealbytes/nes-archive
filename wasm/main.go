package main

import (
	_ "embed"
	"fmt"
	"syscall/js"
	"time"

	"github.com/fogleman/nes/nes"
)

//go:embed state/mario.static
var staticData []byte

//go:embed state/mario.dyn
var dynamicData []byte

const (
	NES_WIDTH  = 256
	NES_HEIGHT = 240
)

func main() {
	kb := NewKeyboard()
	renderer := NewRenderer()

	nes, err := nes.NewHeadlessConsole(staticData, dynamicData)
	if err != nil {
		panic(err)
	}

	spf := time.Second / 60
	ticker := time.NewTicker(spf)
	defer ticker.Stop()

	for range ticker.C {
		controller := kb.getController()
		nes.Controller1.SetButtons(controller)
		startTime := time.Now()
		nes.StepSeconds(spf.Seconds())
		fmt.Println("[wasm] Stepped", spf, "in", time.Since(startTime))
		startTime = time.Now()
		renderer.renderImage(nes.Buffer().Pix)
		fmt.Println("[wasm] Renderer in", time.Since(startTime))
	}
}

type renderer struct {
	document  js.Value
	canvas    js.Value
	context   js.Value
	imageData js.Value
	jsData    js.Value
}

func NewRenderer() *renderer {
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	context := canvas.Call("getContext", "2d")
	imageData := context.Call("createImageData", NES_WIDTH, NES_HEIGHT)
	jsData := imageData.Get("data")
	return &renderer{
		document:  document,
		canvas:    canvas,
		context:   context,
		imageData: imageData,
		jsData:    jsData,
	}
}

func (r *renderer) renderImage(rawImageData []uint8) {
	r.context.Call("clearRect", 0, 0, r.canvas.Get("width").Int(), r.canvas.Get("height").Int())
	js.CopyBytesToJS(r.jsData, rawImageData)
	r.context.Call("putImageData", r.imageData, 0, 0)
}

type keyboard struct {
	keyStates map[string]bool
}

func NewKeyboard() *keyboard {
	kb := &keyboard{make(map[string]bool)}
	window := js.Global().Get("window")
	window.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		key := event.Get("code").String()
		kb.keyStates[key] = true
		return nil
	}))
	window.Call("addEventListener", "keyup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		key := event.Get("code").String()
		kb.keyStates[key] = false
		return nil
	}))
	return kb
}

func (kb *keyboard) getController() [8]bool {
	return [8]bool{
		kb.keyStates["KeyZ"],
		kb.keyStates["KeyX"],
		kb.keyStates["ShiftRight"],
		kb.keyStates["Enter"],
		kb.keyStates["ArrowUp"],
		kb.keyStates["ArrowDown"],
		kb.keyStates["ArrowLeft"],
		kb.keyStates["ArrowRight"],
	}
}
