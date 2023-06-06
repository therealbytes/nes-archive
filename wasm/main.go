package main

import (
	"encoding/json"
	"fmt"
	"image"
	"syscall/js"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fogleman/nes/nes"
)

const (
	NES_WIDTH  = 256
	NES_HEIGHT = 240
)

var preimageCache = map[common.Hash][]byte{}

func main() {

	fmt.Println("[wasm] Initializing")

	kb := NewKeyboard()
	renderer := NewRenderer()
	api := NewAPI()
	governor := NewMovingAverage(60)
	recorder := NewRecorder()

	var machine *nes.Console

	spf := time.Second / 30
	spfMs := int(spf.Milliseconds())
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
			// renderer.dim()
			renderer.canvas.Get("classList").Call("add", "paused")
		pause:
			for {
				select {
				case <-api.pauseChan:
					fmt.Println("[wasm] Already paused")
					continue
				case <-api.unpauseChan:
					fmt.Println("[wasm] Unpausing")
					renderer.canvas.Get("classList").Call("remove", "paused")
					break pause
				}
			}
			fmt.Println("[wasm] Unpaused")
		case <-api.requestActivityChan:
			fmt.Println("[wasm] Requesting activity")
			if machine == nil {
				api.returnHashChan <- common.Hash{}
				api.returnActivityChan <- []Action{}
				continue
			}
			// dyn, err := machine.SerializeDynamic()
			// if err != nil {
			// 	panic(err)
			// }
			// hash := crypto.Keccak256Hash(dyn)
			// preimageCache[hash] = dyn
			// fmt.Println("[wasm] Cached dynamic state", hash.Hex())
			api.returnHashChan <- common.Hash{}
			api.returnActivityChan <- recorder.getActivity()
			fmt.Println("[wasm] Returned activity")
		case newSpeed := <-api.speedChan:
			fmt.Println("[wasm] Changing speed to", newSpeed)
			speed = newSpeed
			fmt.Println("[wasm] Speed changed")
		case preimage := <-api.preimageChan:
			fmt.Println("[wasm] Setting preimage", preimage.hash)
			fmt.Println("[wasm] Preimage length:", len(preimage.data))
			if _, ok := preimageCache[preimage.hash]; ok {
				fmt.Println("[wasm] Preimage already exists")
				continue
			}
			preimageCache[preimage.hash] = preimage.data
			fmt.Println("[wasm] Set preimage")
		case cartridge := <-api.cartridgeChan:
			fmt.Println("[wasm] Loading cartridge", cartridge.static, cartridge.dyn)
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
			fmt.Println("[wasm] Static preimage length:", len(staticData))
			fmt.Println("[wasm] Dynamic preimage length:", len(dynData))
			var err error
			machine, err = nes.NewHeadlessConsole(staticData, dynData, true)
			if err != nil {
				fmt.Println("[wasm] Error loading cartridge:", err)
				continue
			}
			fmt.Println("[wasm] Loaded cartridge")
			fmt.Println("[wasm] Resetting recorder")
			recorder.reset()
			fmt.Println("[wasm] Reset recorder")
		case <-ticker.C:
			if machine == nil {
				continue
			}

			startTime := time.Now()

			controller := kb.getController()
			machine.Controller1.SetButtons(controller)
			targetCycles := int(speed * spf.Seconds() * nes.CPUFrequency)
			execCycles := 0

			for execCycles < targetCycles {
				execCycles += machine.Step()
			}

			recorder.record(controller, uint32(execCycles))
			renderer.renderImage(machine.Buffer())

			elapsedTime := time.Since(startTime)
			avgMs := governor.Add(int(elapsedTime.Milliseconds()))

			if avgMs > spfMs-5 {
				fmt.Println("[wasm] Ticking is taking too long:", int(avgMs), "ms/tick")
				newSpeed := speed * 0.99
				speed = newSpeed
				fmt.Println("[wasm] New speed:", speed)
			} else if speed < 1.0 && avgMs < spfMs-10 {
				fmt.Println("[wasm] Ticking is quick:", int(avgMs), "ms/tick")
				newSpeed := speed * 1.01
				if newSpeed > 1.0 {
					newSpeed = 1.0
				}
				speed = newSpeed
				fmt.Println("[wasm] New speed:", speed)
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
	startChan                chan struct{}
	pauseChan                chan struct{}
	unpauseChan              chan struct{}
	speedChan                chan float64
	preimageChan             chan preimage
	cartridgeChan            chan cartridge
	requestActivityChan      chan struct{}
	returnActivityChan       chan []Action
	requestCachePreimageChan chan struct{}
	returnHashChan           chan common.Hash
}

func NewAPI() *nesApi {
	a := &nesApi{
		startChan:                make(chan struct{}, 64),
		pauseChan:                make(chan struct{}, 64),
		unpauseChan:              make(chan struct{}, 64),
		speedChan:                make(chan float64, 64),
		preimageChan:             make(chan preimage, 64),
		cartridgeChan:            make(chan cartridge, 64),
		requestActivityChan:      make(chan struct{}, 64),
		returnActivityChan:       make(chan []Action, 64),
		requestCachePreimageChan: make(chan struct{}, 64),
		returnHashChan:           make(chan common.Hash, 64),
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
			"getActivity": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				hash, activity := a.getActivity()
				activityJson, err := json.Marshal(struct {
					Hash     common.Hash
					Activity []Action
				}{
					hash,
					activity,
				})
				if err != nil {
					panic(err)
				}
				jsActivity := js.Global().Get("Uint8Array").New(len(activityJson))
				js.CopyBytesToJS(jsActivity, activityJson)
				return jsActivity
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

func (a *nesApi) getActivity() (common.Hash, []Action) {
	a.requestActivityChan <- struct{}{}
	hash := <-a.returnHashChan
	activity := <-a.returnActivityChan
	return hash, activity
}

type renderer struct {
	document  js.Value
	canvas    js.Value
	context   js.Value
	imageData js.Value
	jsData    js.Value
}

func NewCanvas(document js.Value, width int, height int) js.Value {
	canvas := document.Call("createElement", "canvas")
	canvas.Set("width", width)
	canvas.Set("height", height)
	document.Get("body").Call("appendChild", canvas)
	return canvas
}

func NewRenderer() *renderer {
	document := js.Global().Get("document")
	canvas := NewCanvas(document, NES_WIDTH, NES_HEIGHT)
	canvas.Get("classList").Call("add", "nes")
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

func (r *renderer) renderImage(img *image.RGBA) {
	rawImageData := img.Pix
	r.context.Call("clearRect", 0, 0, NES_WIDTH, NES_HEIGHT)
	js.CopyBytesToJS(r.jsData, rawImageData)
	r.context.Call("putImageData", r.imageData, 0, 0)
}

func (r *renderer) dim() {
	r.context.Set("fillStyle", "rgba(255, 255, 255, 0.5)")
	width := r.canvas.Get("width").Float()
	height := r.canvas.Get("height").Float()
	r.context.Call("fillRect", 0, 0, width, height)
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
	index  int
	sum    int
	values []int
}

func NewMovingAverage(size int) *MovingAverage {
	return &MovingAverage{values: make([]int, size)}
}

func (ma *MovingAverage) Add(value int) int {
	ma.sum -= ma.values[ma.index]
	ma.sum += value
	ma.values[ma.index] = value
	ma.index = (ma.index + 1) % len(ma.values)
	return ma.sum / len(ma.values)
}

type Action struct {
	Button   uint8
	Press    bool
	Duration uint32
}

type recorder struct {
	buttons  [8]bool
	activity []Action
}

func NewRecorder() *recorder {
	r := &recorder{
		buttons:  [8]bool{},
		activity: make([]Action, 0),
	}
	r.reset()
	return r
}

func (r *recorder) record(buttons [8]bool, duration uint32) {
	if buttons != r.buttons {
		for button, press := range buttons {
			if press != r.buttons[button] {
				action := Action{uint8(button), press, 0}
				r.activity = append(r.activity, action)
			}
		}
		r.buttons = buttons
	}
	r.activity[len(r.activity)-1].Duration += duration
}

func (r *recorder) getActivity() []Action {
	return r.activity
}

func (r *recorder) reset() {
	r.activity = make([]Action, 0)
	nilAction := Action{0, false, 0}
	r.activity = append(r.activity, nilAction)
}
