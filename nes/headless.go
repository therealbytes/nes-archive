package nes

func NewHeadlessConsole(static []byte, dynamic []byte, renderPixels bool, stepAPU bool) (*Console, error) {
	cartridge := &Cartridge{}
	ram := make([]byte, 2048)
	controller1 := NewController()
	controller2 := NewController()
	meta := &MetaConfig{Headless: true, RenderPixels: renderPixels, StepAPU: stepAPU}
	console := Console{meta, nil, nil, nil, cartridge, controller1, controller2, nil, ram}

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
