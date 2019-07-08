package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// File struct
type File struct {
	SortMode string `json:"sort_mode"`
	Ncurses  bool   `json:"ncurses_mode"`
	Update   bool   `json:"always_update_repos"`
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
func ReadConfigFile() error {
	fileOpen, err := os.Open(files.ConfigFile)
	if err != nil {
		return err
	}

	data, _ := ioutil.ReadAll(fileOpen)

	var file File
	errC := json.Unmarshal(data, &file)
	if errC != nil {
		// Problem parsing file
		// Write new config
		errInit := InitConfig(fileOpen)
		if errInit != nil {
			return errInit
		}
	}

	// Set config
	files.UserFile = file
	return nil
}

// InitConfig writes the initial JSON to the file
func InitConfig(file *os.File) error {
	initFile := &File{
		SortMode: "closest",
		Ncurses:  true,
		Update:   false,
	}
	write, err := json.MarshalIndent(initFile, "", "  ")
	if err != nil {
		return err
	}
	if _, errF := file.Write(write); errF != nil {
		return errF
	}
	return nil
}
