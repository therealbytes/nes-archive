package main

import (
	"fmt"
	"log"
	"os"
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

	static, err := console.SerializeStatic()
	if err != nil {
		log.Fatalln(err)
	}

	dyn, err := console.SerializeDynamic()
	if err != nil {
		log.Fatalln(err)
	}

	// write static data
	staticPath := noext + ".static"
	fmt.Println("Writing static data:", staticPath)
	writeFile(staticPath, static)

	// write dynamic data
	dynPath := noext + ".dyn"
	fmt.Println("Writing dynamic data:", dynPath)
	writeFile(dynPath, dyn)
}

func writeFile(path string, data []byte) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		log.Fatalln(err)
	}
}
