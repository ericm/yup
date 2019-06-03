package search

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ericm/yup/output"
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
	InstalledSize    int64
	SortValue        float64
}

// Pacman returns []Package parsed from pacman
func Pacman(query string) ([]Package, error) {
	search := exec.Command("pacman", "-Ss", query)
	run, err := search.Output()
	if err != nil {
		return []Package{}, output.Errorf("%s", err)
	}

	// Find Package vals
	repoRe := regexp.MustCompile("^([A-z]+)")

	searchOutput := string(run)
	pacOut := []string{}
	last := ""
	for i, pac := range strings.Split(searchOutput, "\n") {
		if i%2 == 0 {
			last = pac
		} else {
			pacOut = append(pacOut, fmt.Sprintf("%s\n%s", last, pac))
		}
	}

	fmt.Println(pacOut[0])

	for _, pac := range pacOut {
		fmt.Println(repoRe.FindString(pac))
	}

	var packs []Package

	return packs, nil
}
