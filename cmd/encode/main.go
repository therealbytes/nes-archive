package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/fogleman/nes/cmd"
	"github.com/fogleman/nes/nes"
)

func main() {
	log.SetFlags(0)
	paths := cmd.GetPaths()
	if len(paths) == 0 {
		log.Fatalln("no rom files specified or found")
	}
	for _, path := range paths {
		encode(path)
	}
}

func encode(path string) {
	fmt.Println("Encoding ROM:", path)
	console, err := nes.NewConsole(path)
	if err != nil {
		log.Fatalln(err)
	}
	noext := strings.TrimSuffix(path, filepath.Ext(path))
	err = console.SaveStateStatic(noext)
	if err != nil {
		log.Fatalln(err)
	}
	err = console.SaveStateDynamic(noext)
	if err != nil {
		log.Fatalln(err)
	}
}
