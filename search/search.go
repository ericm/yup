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
	InstalledSize    string
	DownloadSize     string
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

	// Regex definitions
	repoRe := regexp.MustCompile("^([A-z]+)")
	nameRe := regexp.MustCompile("(?:/)+(\\S+)")
	versionRe := regexp.MustCompile("^(?:\\S+ ){1}(\\S+)")
	installedRe := regexp.MustCompile("\\[(.+)\\]")

	siRe := regexp.MustCompile("(?:\\:)(.+)")

	packs := []Package{}
	for _, pac := range pacOut {
		pack := Package{
			Name:        nameRe.FindString(pac)[1:],
			Repo:        repoRe.FindString(pac),
			Version:     strings.Split(versionRe.FindString(pac), " ")[1],
			Installed:   len(installedRe.FindString(pac)) != 0,
			Description: strings.Split(pac, "\n")[1][4:],
		}
		if pack.Installed {
			// Add extra install info
			pacmanSi := exec.Command("pacman", "-Si", pack.Name)
			siOut, err := pacmanSi.Output()
			if err != nil {
				return []Package{}, output.Errorf("%s", err)
			}
			// Get info from pacman -Si package
			info := siRe.FindAllString(string(siOut), 18)

			pack.InstalledVersion = info[2][2:]
			pack.InstalledSize = info[14][2:]
			pack.DownloadSize = info[13][2:]
		}

		packs = append(packs, pack)
	}

	return packs, nil
}
