package main

import (
	_ "embed"
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
	// kb := NewKeyboard()
	renderer := NewRenderer()

	nes, err := nes.NewHeadlessConsole(staticData, dynamicData)
	if err != nil {
		panic(err)
	}

	spf := time.Second / 10
	ticker := time.NewTicker(spf)
	defer ticker.Stop()

	for range ticker.C {
		nes.StepSeconds(spf.Seconds())
		renderer.renderImage(nes.Buffer().Pix)
	}
}

type renderer struct{}

func NewRenderer() *renderer {
	return &renderer{}
}

func (r *renderer) renderImage(rawImageData []uint8) {
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	context := canvas.Call("getContext", "2d")
	imageData := context.Call("createImageData", NES_WIDTH, NES_HEIGHT)
	jsData := imageData.Get("data")
	js.CopyBytesToJS(jsData, rawImageData)
	context.Call("putImageData", imageData, 0, 0)
}

type keyboard struct {
	keyStates     map[string]bool
	pendingKeyUps map[string]bool
}

func NewKeyboard() *keyboard {
	kb := &keyboard{
		keyStates:     make(map[string]bool),
		pendingKeyUps: make(map[string]bool),
	}
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
		kb.pendingKeyUps[key] = false
		return nil
	}))
	return kb
}

func (kb *keyboard) finaliseKeyUps() {
	for key, value := range kb.pendingKeyUps {
		kb.keyStates[key] = value
	}
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
