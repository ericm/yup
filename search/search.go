package search

import (
	"os/exec"

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
func Pacman(query string) []Package {
	search := exec.Command("pacman", "-Ss", query)
	output.SetStd(search)

	var packs []Package
	// TODO: parse search

	return packs
}
