package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ericm/yup/output"
	"github.com/ericm/yup/search"

	"github.com/ericm/yup/sync"
)

// Arguments represent the args passed
type Arguments struct {
	args         []string
	sendToPacman bool
	sync         bool
	// Map of individual args
	options map[string]bool
	targets []string
}

type pair struct {
	a, b string
}

// Constants for output
const help = `Usage:
    yup`

// Custom commands not to be passed to pacman
var commands []pair
var commandShort map[string]bool
var commandLong map[string]bool

func init() {
	commandShort = make(map[string]bool)
	commandLong = make(map[string]bool)

	// Initial definition of custom commands
	commands = []pair{
		pair{"h", "help"},
		pair{"V", "version"},
		// Handle sync
		pair{"S", "sync"},
	}

	for _, arg := range commands {
		commandShort[arg.a] = true
		commandLong[arg.b] = true
	}
}

var arguments = &Arguments{sendToPacman: false, sync: false, options: make(map[string]bool), targets: []string{}}

// Execute initialises the arguments slice and parses args
func Execute() error {
	arguments.args = append(arguments.args, os.Args[1:]...)
	arguments.genOptions()
	arguments.isPacman()
	if arguments.sendToPacman {
		// send to pacman
		sendToPacman()
	} else {
		return arguments.getActions()
	}
	return nil
}

func sendToPacman() {
	allArgs := append([]string{"pacman"}, arguments.args...)

	pacman := exec.Command("sudo", allArgs...)
	output.SetStd(pacman)
	pacman.Run()
}

// Arguments methods

// Generates arguments.options
func (args *Arguments) genOptions() {
	for _, arg := range args.args {
		if arg[:2] == "--" {
			// Long command
			args.options[arg[2:]] = true
		} else if arg[:1] == "-" {
			// Short command
			for i := 1; i < len(arg); i++ {
				args.options[arg[i:i+1]] = true
			}
		} else {
			// Set targets
			args.targets = append(args.targets, arg)
		}
	}
}

// getActions routes the actions
func (args *Arguments) getActions() error {
	if args.sync {
		if len(args.args) == 0 {
			// Update
		} else {
			// Call search
		}
	} else {
		if args.argExist("h", "help") {
			// Help
			return nil
		}
		if args.argExist("S", "sync") {
			return args.syncCheck()
		}

		if args.argExist("V", "version") {
			// Version
			return nil
		}
	}
	// Probs shouldn't reach this point
	return fmt.Errorf("Error in parsing operations")
}

// isPacman checks if the commands are custom yup commands
func (args *Arguments) isPacman() {
	for _, arg := range args.args {
		if len(arg) > 2 && arg[:2] == "--" {
			args.sendToPacman = !customLong(arg[2:])
			return
		} else if len(arg) > 1 && arg[:1] == "-" {
			args.sendToPacman = !customShort(arg[1:2])
			return
		}
	}
	// No flags passed
	args.sync = true
	args.sendToPacman = false
}

// syncCheck checks -S argument options
func (args *Arguments) syncCheck() error {
	if args.argExist("y", "refresh") {
		// Refresh
		output.Printf("Refreshing local repositories")
		refresh := exec.Command("sudo", "pacman", "-Sy")
		output.SetStd(refresh)
		if err := refresh.Run(); err != nil {
			return err
		}

	}
	if args.argExist("s", "search") {
		// Search
		// Check for q
		_, err := search.Pacman(strings.Join(args.targets, " "))
		return err
	}
	if args.argExist("u", "upgrade") {

	}
	if args.argExist("p", "print") {

	}
	if args.argExist("c", "clean") {

	}
	if args.argExist("l", "list") {

	}
	if args.argExist("i", "info") {

	}

	// Default case
	return sync.Sync(args.targets)
}

// Returns whether or not an arg exists
func (args *Arguments) argExist(keys ...string) bool {
	for _, key := range keys {
		if _, exists := args.options[key]; exists {
			return true
		}
	}
	return false
}

// toString for args
func (args *Arguments) toString() string {
	var str = ""
	for _, arg := range args.args {
		str += " " + arg
	}
	return str[1:]
}

func customLong(arg string) bool {
	_, exists := commandLong[arg]
	return exists
}

func customShort(arg string) bool {
	_, exists := commandShort[arg]
	return exists
}
