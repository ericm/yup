package config

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ericm/yup/output"
)

// File struct
type File struct {
	SortMode      string `json:"sort_mode"`
	Ncurses       bool   `json:"ncurses_mode"`
	Update        bool   `json:"always_update_repos"`
	PrintPkg      bool   `json:"print_pkgbuild"`
	AskPkg        bool   `json:"ask_pkgbuild"`
	AskRedo       bool   `json:"ask_redo"`
	ConfigVersion string `json:"version"`
	SilentUpdate  bool   `json:"silent_update"`
	PacmanLimit   int    `json:"pacman_limit"`
	AurLimit      int    `json:"aur_limit"`
}

// Config struct
type Config struct {
	CacheDir   string
	ConfigDir  string
	ConfigFile string
	Ncurses    bool
	UserFile   File
}

// Files represents the config files / dirs
var files Config

func init() {
	files = Config{}
}

// SetConfig is a setter for config
func SetConfig(config *Config) {
	files = *config
}

// GetConfig is a getter for config
func GetConfig() *Config {
	return &files
}

// ReadConfigFile reads the json config
func ReadConfigFile(version string) error {
	fileOpen, err := os.OpenFile(files.ConfigFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	data, _ := ioutil.ReadAll(fileOpen)

	var file File
	errC := json.Unmarshal(data, &file)
	if errC != nil {
		// Problem parsing file
		// Write new config
		errInit := InitConfig(fileOpen, version)
		if errInit != nil {
			return errInit
		}
		// Return as config verson should match up
		return nil
	}

	if file.ConfigVersion != version {
		// Ask user to remake config file
		output.PrintIn("An update was detected. Remake config file? (y/n)")
		scanner := bufio.NewReader(os.Stdin)
		userIn, _ := scanner.ReadString('\n')
		var up bool
		switch strings.ToLower(strings.TrimSpace(userIn))[0] {
		case 'y':
			up = true
		}

		if up {
			errInit := InitConfig(fileOpen, version)
			if errInit != nil {
				return errInit
			}
		} else {
			file.ConfigVersion = version
			// Write
			write, err := json.MarshalIndent(file, "", "  ")
			if err != nil {
				return err
			}
			if _, errF := fileOpen.WriteAt(write, 0); errF != nil {
				return errF
			}
		}
	}

	// Set config
	files.UserFile = file
	return nil
}

// InitConfig writes the initial JSON to the file
func InitConfig(file *os.File, version string) error {
	initFile := &File{
		SortMode:      "closest",
		Ncurses:       true,
		Update:        false,
		PrintPkg:      true,
		AskPkg:        true,
		AskRedo:       true,
		ConfigVersion: version,
		SilentUpdate:  false,
		PacmanLimit:   20,
		AurLimit:      30,
	}
	write, err := json.MarshalIndent(initFile, "", "  ")
	if err != nil {
		return err
	}
	if _, errF := file.WriteAt(write, 0); errF != nil {
		// Read file if error
		ReadConfigFile(version)
	}
	files.UserFile = *initFile
	return nil
}
