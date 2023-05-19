package nes

import (
	"io/ioutil"
	"testing"
)

func TestHeadless(t *testing.T) {
	static, err := ioutil.ReadFile("../mario.static")
	if err != nil {
		t.Fatal(err)
	}
	dynamic, err := ioutil.ReadFile("../mario.dynamic")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewHeadless(static, dynamic)
	if err != nil {
		t.Fatal(err)
	}
}
