package main

import (
	_ "embed"
	"fmt"
	"syscall/js"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fogleman/nes/nes"
)

const (
	NES_WIDTH  = 256
	NES_HEIGHT = 240
)

// TODO: proper cache
var preimageCache = map[common.Hash][]byte{}

func main() {

	fmt.Println("[wasm] Initializing")

	kb := NewKeyboard()
	renderer := NewRenderer()
	api := NewAPI()
	governor := NewMovingAverage(15)

	var machine *nes.Console

	spf := time.Second / 60
	speed := 1.0

	<-api.startChan
	fmt.Println("[wasm] Starting")

	ticker := time.NewTicker(spf)
	defer ticker.Stop()

	for {
		select {
		case <-api.unpauseChan:
			fmt.Println("[wasm] Already unpaused")
			continue
		case <-api.pauseChan:
			fmt.Println("[wasm] Paused")
			renderer.dim()
		pause:
			for {
				select {
				case <-api.pauseChan:
					fmt.Println("[wasm] Already paused")
					continue
				case <-api.unpauseChan:
					fmt.Println("[wasm] Unpausing")
					break pause
				}
			}
			fmt.Println("[wasm] Unpaused")
		case newSpeed := <-api.speedChan:
			speed = newSpeed
			fmt.Println("[wasm] Speed changed to", newSpeed)
		case preimage := <-api.preimageChan:
			if _, ok := preimageCache[preimage.hash]; ok {
				fmt.Println("[wasm] Preimage already exists")
				continue
			}
			preimageCache[preimage.hash] = preimage.data
			fmt.Println("[wasm] Set preimage", preimage.hash)
		case cartridge := <-api.cartridgeChan:
			staticData, ok := preimageCache[cartridge.static]
			if !ok {
				fmt.Println("[wasm] Static preimage not found")
				continue
			}
			dynData, ok := preimageCache[cartridge.dyn]
			if !ok {
				fmt.Println("[wasm] Dynamic preimage not found")
				continue
			}
			var err error
			machine, err = nes.NewHeadlessConsole(staticData, dynData)
			if err != nil {
				fmt.Println("[wasm] Error loading cartridge:", err)
				continue
			}
			fmt.Println("[wasm] Inserted cartridge", cartridge.static, cartridge.dyn)
		case <-ticker.C:
			startTime := time.Now()
			if machine == nil {
				// fmt.Println("[wasm] No cartridge")
				continue
			}
			steps := int(speed * spf.Seconds() * nes.CPUFrequency)
			controller := kb.getController()
			machine.Controller1.SetButtons(controller)
			for i := 0; i < steps; i++ {
				machine.Step()
			}
			renderer.renderImage(machine.Buffer().Pix)
			elapsedTime := time.Since(startTime)
			avgMs := governor.Add(float64(elapsedTime.Milliseconds()))
			spfMs := float64(spf.Milliseconds())
			if avgMs > spfMs*0.875 {
				newSpeed := speed * 0.99
				speed = newSpeed
			} else if speed < 1.0 && avgMs < spfMs*0.875 {
				newSpeed := speed * 1.005
				if newSpeed > 1.0 {
					newSpeed = 1.0
				}
				speed = newSpeed
			}
		}
	}
}

type preimage struct {
	hash common.Hash
	data []byte
}

type cartridge struct {
	static common.Hash
	dyn    common.Hash
}

type nesApi struct {
	startChan     chan struct{}
	pauseChan     chan struct{}
	unpauseChan   chan struct{}
	speedChan     chan float64
	preimageChan  chan preimage
	cartridgeChan chan cartridge
}

func NewAPI() *nesApi {
	a := &nesApi{
		startChan:     make(chan struct{}, 64),
		pauseChan:     make(chan struct{}, 64),
		unpauseChan:   make(chan struct{}, 64),
		speedChan:     make(chan float64, 64),
		preimageChan:  make(chan preimage, 64),
		cartridgeChan: make(chan cartridge, 64),
	}
	js.Global().Set("NesAPI", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return map[string]interface{}{
			"start": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				a.start()
				return nil
			}),
			"pause": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				a.pause()
				return nil
			}),
			"unpause": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				a.unpause()
				return nil
			}),
			"setSpeed": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				a.setSpeed(args[0].Float())
				return nil
			}),
			"setPreimage": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				jsHash := args[0]
				jsData := args[1]

				hash := make([]byte, jsHash.Get("byteLength").Int())
				data := make([]byte, jsData.Get("byteLength").Int())

				js.CopyBytesToGo(hash, jsHash)
				js.CopyBytesToGo(data, jsData)

				hashInGo := common.BytesToHash(hash)
				a.setPreimage(hashInGo, data)
				return nil
			}),
			"setCartridge": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				jsStatic := args[0]
				jsDyn := args[1]

				static := make([]byte, jsStatic.Get("byteLength").Int())
				dyn := make([]byte, jsDyn.Get("byteLength").Int())

				js.CopyBytesToGo(static, jsStatic)
				js.CopyBytesToGo(dyn, jsDyn)

				staticInGo := common.BytesToHash(static)
				dynInGo := common.BytesToHash(dyn)
				a.setCartridge(staticInGo, dynInGo)
				return nil
			}),
		}
	}))
	return a
}

func (a *nesApi) start() {
	a.startChan <- struct{}{}
}

func (a *nesApi) pause() {
	a.pauseChan <- struct{}{}
}

func (a *nesApi) unpause() {
	a.unpauseChan <- struct{}{}
}

func (a *nesApi) setSpeed(speed float64) {
	a.speedChan <- speed
}

func (a *nesApi) setPreimage(hash common.Hash, data []byte) {
	a.preimageChan <- preimage{hash, data}
}

func (a *nesApi) setCartridge(static, dyn common.Hash) {
	a.cartridgeChan <- cartridge{static, dyn}
}

type renderer struct {
	document  js.Value
	canvas    js.Value
	context   js.Value
	imageData js.Value
	jsData    js.Value
}

func NewCanvas(document js.Value) js.Value {
	// Create a new canvas element
	canvas := document.Call("createElement", "canvas")

	// Set the canvas id
	// canvas.Set("id", "myCanvas")

	// Set canvas width and height
	canvas.Set("width", NES_WIDTH)
	canvas.Set("height", NES_HEIGHT)
	canvas.Get("classList").Call("add", "nes")

	// Append the canvas to the body of the document
	document.Get("body").Call("appendChild", canvas)

	return canvas
}

func NewRenderer() *renderer {
	document := js.Global().Get("document")
	canvas := NewCanvas(document)
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

func (r *renderer) dim() {
	r.context.Set("fillStyle", "rgba(255, 255, 255, 0.5)")
	width := r.canvas.Get("width").Float()
	height := r.canvas.Get("height").Float()
	r.context.Call("fillRect", 0, 0, width, height)
}

// TODO: set controller t/ API [?]

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

func (kb *keyboard) isPressed(key string) bool {
	return kb.keyStates[key]
}

func (kb *keyboard) getController() [8]bool {
	return [8]bool{
		kb.isPressed("KeyZ"),
		kb.isPressed("KeyX"),
		kb.isPressed("ShiftRight"),
		kb.isPressed("Enter"),
		kb.isPressed("ArrowUp"),
		kb.isPressed("ArrowDown"),
		kb.isPressed("ArrowLeft"),
		kb.isPressed("ArrowRight"),
	}
}

type MovingAverage struct {
	size  int
	sum   float64
	queue []float64
}

func NewMovingAverage(size int) *MovingAverage {
	return &MovingAverage{
		size:  size,
		queue: make([]float64, 0, size),
	}
}

func (ma *MovingAverage) Add(value float64) float64 {
	if len(ma.queue) >= ma.size {
		ma.sum -= ma.queue[0]
		ma.queue = ma.queue[1:]
	}
	ma.queue = append(ma.queue, value)
	ma.sum += value
	return ma.sum / float64(len(ma.queue))
}
