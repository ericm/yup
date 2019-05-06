package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const configFileName string = "yay.conf"

func main() {
	if os.Geteuid() == 0 {
		fmt.Fprintln(os.Stderr, "No need to run yup as sudo/root")
	}
	if runtime.GOOS != "linux" {
		fmt.Fprintln(os.Stderr, "Requires linux I'm afraid")
		os.Exit(1)
	}
	exitError(paths())

}

func exitError(err error) {
	if err != nil {
		if str := err.Error(); str != "" {
			fmt.Fprintln(os.Stderr, str)
			os.Exit(1)
		}
	}
}

func paths() error {
	configDir := os.Getenv("HOME")
	cacheDir := configDir
	filepath.Join(configDir, ".config/yup")
	filepath.Join(cacheDir, ".cache/yup")
	return nil
}

func makeConfig() {

}
