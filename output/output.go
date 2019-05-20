package output

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	// ARROW (purple) for printing
	ARROW string = "\033[95m==>\033[0m"
	// ARROWERROR for error
	ARROWERROR string = "\033[31m==>\033[0m"
)

// Printf arrow wrapper for fmt
func Printf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", ARROW, fmt.Sprintf(format, a...))
}

// Errorf arrow wrapper for fmt
func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf("%s %s\n", ARROWERROR, fmt.Sprintf(format, a...))
}

// PrintL - prints line break
func PrintL() {
	fmt.Printf("\033[95m- - - - - -\033[0m\n")
}

// SetStd sets cmd's Stdout, Stderr and Stdin to the OS's
func SetStd(cmd *exec.Cmd) {
	cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
}
