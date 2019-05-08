package cmd

import (
	"fmt"
	"os"
	"os/exec"

	mapset "github.com/deckarep/golang-set"
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

// Custom commands not to be passed to pacman
var commands []pair
var commandShort mapset.Set
var commandLong mapset.Set

func init() {
	commandShort = mapset.NewSet()
	commandLong = mapset.NewSet()

	// Initial definition of custom commands
	commands = []pair{
		pair{"h", "help"},
		pair{"v", "version"},
		// Handle sync
	}

	for _, arg := range commands {
		commandShort.Add(arg.a)
		commandLong.Add(arg.b)
	}
}

// Routes actions for custom commands
func action(arg string) {
	switch arg {
	case "h":
		fmt.Println(`Usage:
    yay`)
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
		allArgs := append([]string{"pacman"}, arguments.args...)

		pacman := exec.Command("sudo", allArgs...)
		pacman.Stdout, pacman.Stdin, pacman.Stderr = os.Stdout, os.Stdin, os.Stderr
		pacman.Run()
	} else {
		arguments.getActions()
	}
	return nil
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

// toString for args
func (args *Arguments) toString() string {
	var str = ""
	for _, arg := range args.args {
		str += " " + arg
	}
	return str[1:]
}

func customLong(arg string) bool {
	return commandLong.Contains(arg)
}

func customShort(arg string) bool {
	return commandShort.Contains(arg)
}
