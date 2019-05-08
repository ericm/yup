package cmd

import (
	"fmt"
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
	commandShort = mapset.NewSet()
	commandLong = mapset.NewSet()

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
	arguments.args = append(arguments.args, os.Args[1:]...)
	arguments.isPacman()
	if arguments.sendToPacman {
		// send to pacman
		fmt.Println(arguments.args)
	} else {

	}
	return nil
}

// Arguments methods

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
	args.sendToPacman = false
}

func customLong(arg string) bool {
	return commandLong.Contains(arg)
}

func customShort(arg string) bool {
	return commandShort.Contains(arg)
}
