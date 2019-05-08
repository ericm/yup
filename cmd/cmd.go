package cmd

import (
	"os"

	mapset "github.com/deckarep/golang-set"
)

// Arguments represent the args passed
type Arguments struct {
	args         []string
	sendToPacman bool
}

type pair struct {
	a, b string
}

// Custom commands not to be passed to pacman
var commands []pair
var commandShort mapset.Set
var commandLong mapset.Set

func init() {
	commands = []pair{
		pair{"h", "help"},
		pair{"v", "version"},
	}
	for _, arg := range commands {
		commandShort.Add(arg.a)
		commandLong.Add(arg.b)
	}
}

var arguments = &Arguments{sendToPacman: false}

// Execute initialises the arguments slice and parses args
func Execute() error {
	arguments.args = append(arguments.args, os.Args...)
	arguments.isPacman()
	if arguments.sendToPacman {
		// send to pacman
	} else {

	}
	return nil
}

// Arguments methods
func (args *Arguments) isPacman() {
	check := false
	for _, arg := range os.Args {
		if check {
			break
		}
		if len(arg) > 2 && arg[:2] == "--" {
			check = customLong(arg)
		} else if len(arg) > 1 && arg[:1] == "-" {
			check = customShort(arg)
		}
	}
	args.sendToPacman = !check
}

func customLong(arg string) bool {
	return commandLong.Contains(arg[2:])
}

func customShort(arg string) bool {
	return commandShort.Contains(arg[1:])
}
