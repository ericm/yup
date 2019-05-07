package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ericm/yup/cmd"
)

const configFileName = "yay.conf"

// Global variables
var (
	configDir  string
	cacheDir   string
	configFile string
)

// Entrypoint to the program.
// Runs intitialization and calls other packages.
func main() {
	if os.Geteuid() == 0 {
		fmt.Fprintln(os.Stderr, "No need to run yup as sudo/root")
	}
	if runtime.GOOS != "linux" {
		fmt.Fprintln(os.Stderr, "Requires linux I'm afraid")
		os.Exit(1)
	}

	exitError(paths())
	exitError(makePaths())
	exitError(cmd.Execute())
}

// Exits the program upon a core function returning an error.
func exitError(err error) {
	if err != nil {
		if str := err.Error(); str != "" {
			fmt.Fprintln(os.Stderr, str)
			os.Exit(1)
		}
	}
}

// Sets global path variables
func paths() error {
	configDir = os.Getenv("HOME")
	if len(configDir) == 0 {
		return fmt.Errorf("HOME environment variable unset")
	}

	cacheDir = filepath.Join(configDir, ".cache/yup")
	configDir = filepath.Join(configDir, ".config/yup")

	configFile = filepath.Join(configDir, configFileName)
	return nil
}

// Checks and creates the directories and file required by yup.
func makePaths() error {
	// Config
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err = os.MkdirAll(configDir, 0755); err != nil {
			fmt.Println(len(configDir))
			return fmt.Errorf("Failed to create config directory: %s", err)
		}
		// Create file
		os.OpenFile(configFile, os.O_CREATE, 0664)
	} else if err != nil {
		return err
	}

	// Cache
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err = os.MkdirAll(cacheDir, 0755); err != nil {
			return fmt.Errorf("Failed to create cache directory: %s", err)
		}
	} else if err != nil {
		return err
	}

	return nil
}
