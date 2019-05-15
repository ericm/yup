package output

import (
	"fmt"
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
