package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ericm/yup/config"

	"github.com/ericm/yup/cmd"
	"github.com/ericm/yup/output"
)

const configFileName = "config.json"

// Global variables
var (
	configDir  string
	cacheDir   string
	configFile string
	configSet  *config.Config
)

// Entrypoint to the program.
// Runs intitialization and calls other packages.
func main() {
	if os.Geteuid() == 0 {
		output.PrintErr("No need to run yup as sudo/root")
		os.Exit(1)
	}
	if runtime.GOOS != "linux" {
		fmt.Fprintln(os.Stderr, "Requires linux I'm afraid")
		os.Exit(1)
	}

	exitError(paths())
	exitError(makePaths())
	exitError(config.ReadConfigFile(cmd.Version))
	exitError(cmd.Execute())
}

// Exits the program upon a core function returning an error.
func exitError(err error) {
	if err != nil {
		if str := err.Error(); str != "" {
			fmt.Fprintln(os.Stderr, output.Errorf(str))
			os.Exit(1)
		}
	}
}

// Sets global path variables
func paths() error {
	configSet = &config.Config{}

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
	// Set config
	configSet.ConfigDir = configDir
	configSet.ConfigFile = configFile
	// Config
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err = os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("Failed to create config directory: %s", err)
		}

	} else if err != nil {
		return err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create file
		file, err := os.OpenFile(configFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {

		}
		if err := config.InitConfig(file, cmd.Version); err != nil {
			return err
		}
		if err != nil {
			return err
		}
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
	configSet.CacheDir = cacheDir
	config.SetConfig(configSet)

	return nil
}
