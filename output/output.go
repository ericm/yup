package output

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	// ARROW (purple) for printing
	ARROW string = "\033[95m==>\033[0m"
	// ARROWIN (green) for printing
	ARROWIN string = "\033[92m==>\033[0m"
	// ARROWERROR for error
	ARROWERROR string = "\033[31m==>\033[0m"
)

// Package represents a package in pacman or the AUR
type Package struct {
	Aur              bool
	Repo             string
	Name             string
	Version          string
	Description      string
	Size             int64
	Installed        bool
	InstalledVersion string
	InstalledSize    string
	DownloadSize     string
	InstalledSizeInt int
	SortValue        float64
	OutOfDate        int
}

// Printf arrow wrapper for fmt
func Printf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", ARROW, fmt.Sprintf(format, a...))
}

// PrintIn styles stdout for input from stdin
func PrintIn(format string, a ...interface{}) {
	fmt.Printf("%s \033[92m%s:\033[0m ", ARROWIN, fmt.Sprintf(format, a...))
}

// Errorf arrow wrapper for fmt
func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf("%s %s", ARROWERROR, fmt.Sprintf(format, a...))
}

// PrintErr prints Errorf
func PrintErr(format string, a ...interface{}) {
	fmt.Println(Errorf(format, a...))
}

// PrintL - prints line break
func PrintL() {
	// Get tty size
	tty := exec.Command("stty", "size")
	tty.Stdin = os.Stdin
	out, _ := tty.Output()
	str := strings.Split(string(out), " ")
	var n int
	if len(str) > 1 {
		n, _ = strconv.Atoi(strings.ReplaceAll(str[1], "\n", ""))
	} else {
		n = 40
	}

	fmt.Printf("\033[34m%s\033[0m\n", strings.Repeat("=", n))
}

// SetStd sets cmd's Stdout, Stderr and Stdin to the OS's
func SetStd(cmd *exec.Cmd) {
	cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
}

// PrintPackage in formatted view
func PrintPackage(pack Package, mode ...string) string {
	outdated := ""
	if pack.Version != pack.InstalledVersion {
		outdated = fmt.Sprintf(", (\033[1m\033[95mOUTDATED\033[0m %s)", pack.InstalledVersion)
	}

	out := ""

	if len(mode) > 0 {
		switch mode[0] {
		case "sso":
			// yup -Sso mode
			fmt.Printf("\033[2m/\033[0m\033[1m%s\033[0m %s, (\033[95m\033[1mInstall Size: %s\033[0m)%s\n    %s\n",
				pack.Name, pack.Version, pack.InstalledSize, outdated, pack.Description)
			return ""
		case "def":
		case "ncurses":
			pack.Description = fmt.Sprintf(" - %s", pack.Description)
		}
	}

	if pack.Installed {
		if pack.InstalledSize == "" {
			out = fmt.Sprintf("%s\033[2m/\033[0m\033[1m%s\033[0m %s (\033[1m\033[95mINSTALLED\033[0m)%s\n    %s\n",
				pack.Repo, pack.Name, pack.Version, outdated, pack.Description)
		} else {
			out = fmt.Sprintf("%s\033[2m/\033[0m\033[1m%s\033[0m %s (\033[1m\033[95mINSTALLED\033[0m), (Installed Size: %s)%s\n    %s\n",
				pack.Repo, pack.Name, pack.Version, pack.InstalledSize, outdated, pack.Description)
		}
	} else {
		out = fmt.Sprintf("%s\033[2m/\033[0m\033[1m%s\033[0m %s\n    %s\n", pack.Repo, pack.Name, pack.Version, pack.Description)
	}

	// Handle ncurses
	if len(mode) > 0 && mode[0] == "ncurses" {
		return out
	} else {
		fmt.Print(out)
	}

	return ""
}
