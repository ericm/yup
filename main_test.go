package main

import (
	"testing"
)

func TestPaths(t *testing.T) {
	paths()
	if len(cacheDir) == 0 {
		t.Error("Didnt set cacheDir")
	}
	if len(configDir) == 0 {
		t.Error("Didnt set configDir")
	}
	if len(configFile) == 0 {
		t.Error("Didnt set configFile")
	}
}
