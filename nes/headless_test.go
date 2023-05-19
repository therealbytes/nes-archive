package nes

import (
	"io/ioutil"
	"testing"
)

func TestHeadless(t *testing.T) {
	static, err := ioutil.ReadFile("../roms/mario.static")
	if err != nil {
		t.Fatal(err)
	}
	dynamic, err := ioutil.ReadFile("../roms/mario.dyn")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewHeadless(static, dynamic)
	if err != nil {
		t.Fatal(err)
	}
}
