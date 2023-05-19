package main

import (
	"log"

	"github.com/fogleman/nes/cmd"
	"github.com/fogleman/nes/ui"
)

func main() {
	log.SetFlags(0)
	paths := cmd.GetPaths()
	if len(paths) == 0 {
		log.Fatalln("no rom files specified or found")
	}
	ui.Run(paths)
}
