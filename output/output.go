package output

import (
	"fmt"
)

const (
	// ARROW (purple) for printing
	ARROW string = "\033[95m==>\033[0m"
)

// Printf arrow wrapper for fmt
func Printf(format string, a ...interface{}) {
	fmt.Printf("%s %s", ARROW, fmt.Sprintf(format, a...))
}
