package output

import (
	"fmt"
	"os"
	"os/exec"
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
}

// Printf arrow wrapper for fmt
func Printf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", ARROW, fmt.Sprintf(format, a...))
}

// PrintIn styles stdout for input from stdin
func PrintIn(format string, a ...interface{}) {
	fmt.Printf("%s \033[92m%s:\033[0m", ARROWIN, fmt.Sprintf(format, a...))
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
	fmt.Printf("\033[95m- - - - - -\033[0m\n")
}

// SetStd sets cmd's Stdout, Stderr and Stdin to the OS's
func SetStd(cmd *exec.Cmd) {
	cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
}

// PrintPackage in formatted view
func PrintPackage(pack Package, mode ...string) {
	if len(mode) > 0 {
		switch mode[0] {
		case "sso":
			// yup -Sso mode
			fmt.Printf("%s\033[2m/\033[0m\033[1m%s\033[0m %s, Size: (D: %s | \033[95m\033[1mI: %s\033[0m)\n    %s\n",
				pack.Repo, pack.Name, pack.Version, pack.DownloadSize, pack.InstalledSize, pack.Description)
			return
		}
	} else {
		if pack.Installed {
			if pack.DownloadSize == "" {
				fmt.Printf("%s\033[2m/\033[0m\033[1m%s\033[0m %s (\033[1m\033[95mINSTALLED\033[0m)\n    %s\n",
					pack.Repo, pack.Name, pack.Version, pack.Description)
			} else {
				fmt.Printf("%s\033[2m/\033[0m\033[1m%s\033[0m %s (\033[1m\033[95mINSTALLED\033[0m), Size: (D: %s | I: %s)\n    %s\n",
					pack.Repo, pack.Name, pack.Version, pack.DownloadSize, pack.InstalledSize, pack.Description)
			}

		} else {
			fmt.Printf("%s\033[2m/\033[0m\033[1m%s\033[0m %s\n    %s\n", pack.Repo, pack.Name, pack.Version, pack.Description)
		}
	}

}
