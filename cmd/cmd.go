package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

// Arguments represent the args passed
type Arguments struct {
	args         []string
	sendToPacman bool
	sync         bool
}

type pair struct {
	a, b string
}

// Constants for output
const help = `Usage:
    yay`

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
		pair{"v", "version"},
		// Handle sync
		pair{"S", "sync"},
	}

	for _, arg := range commands {
		commandShort[arg.a] = true
		commandLong[arg.b] = true
	}
}

// Routes actions for custom commands
func action(arg string) {
	switch arg {
	case "h":
		fmt.Println(help)
		break
	case "S":
		// Handle Sync
		arguments.syncCheck()
		break
	}
}

var arguments = &Arguments{sendToPacman: false, sync: false}

// Execute initialises the arguments slice and parses args
func Execute() error {
	arguments.args = append(arguments.args, os.Args[1:]...)
	arguments.isPacman()
	if arguments.sendToPacman {
		// send to pacman
		sendToPacman()
	} else {
		arguments.getActions()
	}
	return nil
}

func sendToPacman() {
	allArgs := append([]string{"pacman"}, arguments.args...)

	pacman := exec.Command("sudo", allArgs...)
	pacman.Stdout, pacman.Stdin, pacman.Stderr = os.Stdout, os.Stdin, os.Stderr
	pacman.Run()
}

// Arguments methods

// getActions routes the actions
func (args *Arguments) getActions() {
	if args.sync {
		if len(args.args) == 0 {
			// Update
		} else {
			// Call search
		}
	} else {
		if len(args.args[0]) == 2 {
			// Call action router if short command
			action(args.args[0][1:])
		} else {
			// Else find the custom command
			// One should always be found due to previous checks
			for _, command := range commands {
				if args.args[0][2:] == command.b {
					action(command.a)
					return
				}
			}
		}
	}
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
	args.sync = true
	args.sendToPacman = false
}

// syncCheck checks -S argument options
func (args *Arguments) syncCheck() {
	if args.argExist('s') {
		// search
	}
	// for _, r := range flag {
	// 	switch r {
	// 	case 's':
	// 		// search
	// 		break
	// 	case 'p':
	// 		// print
	// 		break
	// 	case 'c':
	// 		// clean
	// 		break
	// 	case 'l':
	// 		// list
	// 		break
	// 	case 'i':
	// 		// info
	// 		break
	// 	case 'u':
	// 		// system upgrade
	// 		break
	// 	default:
	// 		sendToPacman()
	// 	}
	// }
}

// Returns whether or not an arg exists
func (args *Arguments) argExist(keys ...rune) bool {
	for _, key := range keys {
		for _, r := range args.args[0] {
			if r == key {
				return true
			}
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
