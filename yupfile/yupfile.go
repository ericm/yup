package yupfile

import (
	"fmt"
	"strings"
)

func Parse(arg string) error {
	args := strings.Split(arg, " ")
	if len(args) > 0 {

		return nil
	}

	return fmt.Errorf("Error parsing yupfile path")
}
