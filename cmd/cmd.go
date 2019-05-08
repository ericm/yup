package cmd

import (
	"os"

	"github.com/deckarep/golang-set"
)

// Arguments represent the args passed
type Arguments struct {
	args         []string
	sendToPacman bool
}

type Pair struct {
	a, b string
}

// Custom commands not to be passed to pacman
var commands []Pair
var commandShort mapset.Set
var commandLong mapset.Set

func init() {
	commands = []Pair{
		Pair{"h", "help"},
		Pair{"v", "version"},
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
	for _, arg := range os.Args {
		if len(arg) > 1 {

		}
	}
	return nil
}

func (args *Arguments) isPacman() {
	args.
}
