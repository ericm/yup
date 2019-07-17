package yupfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Parse(argc string) error {
	args := strings.Split(argc, " ")
	if len(args) > 0 {
		arg := args[0]
		var file *os.File
		var err error
		defer file.Close()
		if filepath.IsAbs(arg) {
			// Absolute path
			file, err = os.OpenFile(arg, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			// Relative path
			dir, _ := os.Getwd()
			os.Chdir(dir)
			file, err = os.OpenFile(arg, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return err
			}
		}
		b, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		data := string(b)
		fmt.Print(data)

		return nil
	}

	return fmt.Errorf("Error parsing yupfile path")
}
