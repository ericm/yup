package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

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
	target  string
}

type pair struct {
	a, b string
}

// Constants for output
const help = `Usage:
  yup
`

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
		pair{"S", "sync"},
		pair{"Q", "query"},
	}

	for _, arg := range commands {
		commandShort[arg.a] = true
		commandLong[arg.b] = true
	}
}

var arguments = &Arguments{sendToPacman: false, sync: false, options: make(map[string]bool), target: ""}

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
		if len(arg) > 1 && arg[:2] == "--" {
			// Long command
			args.options[arg[2:]] = true
		} else if arg[:1] == "-" {
			// Short command
			for i := 1; i < len(arg); i++ {
				args.options[arg[i:i+1]] = true
			}
		} else {
			// Set targets
			if len(args.target) > 0 {
				args.target = fmt.Sprintf("%s %s", args.target, arg)
			} else {
				args.target = arg
			}
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
			output.Printf("Searching and sorting your query...")
			// Get aur packs
			aur, errAur := search.Aur(args.target, false, false)
			if errAur != nil {
				return errAur
			}

			// Get pacman packs
			pacman, errPac := search.Pacman(args.target, false, false)
			if errPac != nil {
				return errPac
			}

			search.SortPacks(args.target, aur, pacman)
			return nil
		}
	} else {
		if args.argExist("h", "help") {
			// Help
			fmt.Print(help)
			return nil
		}

		if args.argExist("S", "sync") {
			return args.syncCheck()
		}

		if args.argExist("V", "version") {
			// Version
			return nil
		}

		if args.argExist("Q", "query") {
			// Check for custom flag; o
			// This sorts by Install size
			if args.argExist("o", "order-by-size") {
				output.Printf("Sorting your query by install size")
				pacman, err := search.PacmanQi()
				sort.Sort(bySize(pacman))

				// Print sorted
				for i, pack := range pacman {
					fmt.Print(len(pacman) - i)
					output.PrintPackage(pack, "sso")
				}

				return err
			}

			// Default case
			sendToPacman()
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

// Soring for packages
// By size
type bySize []output.Package

func (s bySize) Len() int           { return len(s) }
func (s bySize) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s bySize) Less(i, j int) bool { return s[i].InstalledSizeInt < s[j].InstalledSizeInt }

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
		if len(args.target) == 0 {
			return nil
		}

	}
	if args.argExist("s", "search") {
		// Search
		// Check for q
		if args.argExist("q") {
			return nil
		}

		// Only check Aur with a search query
		if len(args.target) > 0 {
			_, errA := search.Aur(args.target, true, false)
			if errA != nil {
				return errA
			}
		}

		_, err := search.Pacman(args.target, true, false)
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
	return sync.Sync([]string{args.target})
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
