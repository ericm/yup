package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/ericm/yup/output"
	"github.com/ericm/yup/search"

	"github.com/ericm/yup/clean"
	"github.com/ericm/yup/config"
	"github.com/ericm/yup/sync"
	"github.com/ericm/yup/update"
	"github.com/ericm/yup/yupfile"
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

// Version of yup
const Version = "1.1.5"

// Constants for output
const help = `Usage:
    yup                 Updates AUR and pacman packages (Like -Syu)
    yup <package(s)>    Searches for that packages and provides an install dialogue

Operations:
    yup {-h --help}             
    yup {-V --version}          
    yup {-D --database} <options> <package(s)>
    yup {-F --files}    <options> <package(s)>
    yup {-Q --query}    <options> <package(s)>
    yup {-R --remove}   <options> <package(s)>
    yup {-S --sync}     <options> <package(s)>
    yup {-T --deptest}  <options> <package(s)>
    yup {-U --upgrade}  <options> <file(s)>

Custom operations:
    yup -c              Cleans cache and unused dependencies
    yup -C              Cleans AUR cache only
    yup -a [package(s)] Operates on the AUR exclusively
    yup -n [package(s)] Runs in non-ncurses mode
    yup -Y <Yupfile>    Install packages from a Yupfile
    yup -Qos            Orders installed packages by install size
`

// Custom commands not to be passed to pacman
var (
	commands     []pair
	commandShort map[string]bool
	commandLong  map[string]bool
)

func init() {
	commandShort = make(map[string]bool)
	commandLong = make(map[string]bool)

	// Initial definition of custom commands
	commands = []pair{
		{"h", "help"},
		{"V", "version"},
		{"S", "sync"},
		{"R", "remove"},
		{"Q", "query"},
		{"c", "clean"},
		{"C", "cache"},
		{"Y", "yupfile"},
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
		sendToPacman(true)
	} else {
		return arguments.getActions()
	}
	return nil
}

func sendToPacman(sudo bool) {
	allArgs := append([]string{"pacman"}, arguments.args...)

	var pacman *exec.Cmd
	if sudo {
		pacman = exec.Command("sudo", allArgs...)
	} else {
		pacman = exec.Command(allArgs[0], allArgs[1:]...)
	}
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
		if len(args.args) == 0 || (len(args.target) == 0 && args.argExist("a", "aur")) {
			// Update
			if args.argExist("a", "aur") {
				return update.AurUpdate()
			}
			return update.Update()
		}
		conFile := config.GetConfig()
		conFile.Ncurses = args.argExist("n", "non-ncurses")
		// Update if wanted
		if conFile.UserFile.Update {
			// Refresh
			output.Printf("Refreshing local repositories")
			refresh := exec.Command("sudo", "pacman", "-Sy")
			output.SetStd(refresh)
			if err := refresh.Run(); err != nil {
				return err
			}
		}
		// Call search
		output.Printf("Searching and sorting your query...")

		// Get aur packs in a gorroutine
		aurChan := make(chan []output.Package)
		var aurErr error
		go func(ch chan []output.Package) {
			aur, errAur := search.Aur(args.target, false, false)
			if errAur != nil {
				aurErr = errAur
			}
			ch <- aur
		}(aurChan)

		// Get pacman packs in a gorroutine
		pacChan := make(chan []output.Package)
		var pacErr error
		if !args.argExist("a", "aur") {
			go func(ch chan []output.Package) {
				// Get pacman packs
				pacman, errPac := search.Pacman(args.target, false, false)
				if errPac != nil {
					pacErr = errPac
				}
				ch <- pacman
			}(pacChan)
		} else {
			go func(ch chan []output.Package) {
				pacChan <- []output.Package{}
			}(pacChan)
		}

		// Combine into one slice
		var packs []output.Package

		for i := 0; i < 2; i++ {
			select {
			case aurPacks := <-aurChan:
				if aurErr != nil {
					output.PrintErr("AUR lookup error: %s", aurErr)
				}
				packs = append(packs, aurPacks...)
			case pacPacks := <-pacChan:
				if pacErr != nil {
					return pacErr
				}
				packs = append(packs, pacPacks...)
			}
		}

		groups, err := search.PacmanGroups(args.target)
		if err != nil {
			return err
		}
		packs = append(packs, groups...)

		search.SortPacks(args.target, packs)
		return nil
	}
	if args.argExist("R", "remove") {
		pkgs := strings.Split(strings.Trim(args.target, " "), " ")
		for _, name := range pkgs {
			if err := sync.Remove(name); err != nil {
				return err
			}
		}
		return nil
	}
	if args.argExist("C", "cache") {
		return clean.Aur()
	}

	if args.argExist("c", "clean") {
		return clean.Clean()
	}

	if args.argExist("Y", "yupfile") {
		return yupfile.Parse(args.target)
	}

	if args.argExist("S", "sync") {
		return args.syncCheck()
	}

	if args.argExist("V", "version") {
		// Version
		fmt.Println(Version)
		return nil
	}

	if args.argExist("Q", "query") {
		// Check for custom flag; -Qos
		// This sorts by Install size
		if args.argExist("o", "order-by") && args.argExist("s", "size") {
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
		sendToPacman(false)
		return nil
	}

	if args.argExist("h", "help") {
		// Help
		fmt.Print(help)
		return nil
	}
	// Probs shouldn't reach this point
	return fmt.Errorf("Error in parsing operations")
}

// isPacman checks if the commands are custom yup commands
func (args *Arguments) isPacman() {
	if args.argExist("a", "aur", "n", "non-ncurses") {
		// Custom args
		args.sync = true
		args.sendToPacman = false
		return
	}

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

// Sorting for packages
// By size
type bySize []output.Package

func (s bySize) Len() int           { return len(s) }
func (s bySize) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s bySize) Less(i, j int) bool { return s[i].InstalledSizeInt < s[j].InstalledSizeInt }

// syncCheck checks -S argument options
func (args *Arguments) syncCheck() error {
	if args.argExist("h", "help", "i", "info", "l", "list", "g", "groups") {
		sendToPacman(false)
		return nil
	}

	if args.argExist("c", "clean", "d", "nodeps", "r", "root") {
		sendToPacman(true)
		return nil
	}

	if args.argExist("y", "refresh") {
		if args.argExist("u", "upgrade") {
			// Upgrade
			return update.Update()
		}
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
		if args.argExist("q", "quiet") {
			return nil
		}

		// Only check Aur with a search query
		if len(args.target) > 0 {
			_, errA := search.Aur(args.target, true, false)
			if errA != nil {
				defer output.PrintErr("AUR query error: %s", errA)
				defer output.PrintL()
			}
		}

		_, err := search.Pacman(args.target, true, false)
		return err
	}

	// Default case
	return sync.Sync(strings.Split(args.target, " "), true, false)
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
