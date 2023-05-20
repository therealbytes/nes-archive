package nes

func NewHeadlessConsole(static []byte, dynamic []byte) (*Console, error) {
	cartridge := &Cartridge{}
	ram := make([]byte, 2048)
	controller1 := NewController()
	controller2 := NewController()
	console := Console{nil, nil, nil, cartridge, controller1, controller2, nil, ram}

	if err := console.DeserializeStatic(static); err != nil {
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

	if err := console.DeserializeDynamic(dynamic); err != nil {
		return nil, err
	}

	return &console, nil
}

// type Action struct {
// 	Button int
// 	Press  bool
// 	Wait   int
// }

// func HeadlessRun(console *Console, activity []Action) {
// 	for _, action := range activity {
// 		if action.Button < 8 {
// 			console.Controller1.buttons[action.Button] = action.Press
// 		}
// 		// Run Wait instructions (not cycles!)
// 		for ii := 0; ii < action.Wait; ii++ {
// 			console.Step()
// 		}
// 	}
// }
