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
func Pacman(query string, print bool) ([]Package, error) {

	search := exec.Command("pacman", "-Ss", query)
	run, err := search.Output()
	if err != nil {
		return []Package{}, err
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
	nameRe := regexp.MustCompile("(?:/)+(\\S+)")
	repoRe := regexp.MustCompile("^([A-z]+)")
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

		// Set colour for repo
		switch pack.Repo {
		case "core":
			// Purple
			pack.Repo = "\033[95mcore\033[0m"
			break
		case "extra":
			// Green
			pack.Repo = "\033[32mextra\033[0m"
			break
		case "community":
			// Cyan
			pack.Repo = "\033[36mcommunity\033[0m"
			break
		case "multilib":
			// Yellow
			pack.Repo = "\033[33mmultilib\033[0m"
			break
		}

		if pack.Installed && len(query) > 0 {
			// Add extra install info
			// Get info from pacman -Sii package
			// Add extra install info
			pacmanSi := exec.Command("pacman", "-Sii", pack.Name)
			siOut, err := pacmanSi.Output()
			if err != nil {
				return []Package{}, output.Errorf("%s", err)
			}

			// Sets the other vals
			info := siRe.FindAllString(string(siOut), -1)
			pack.InstalledVersion = info[2][2:]
			pack.InstalledSize = info[16][2:]
			pack.DownloadSize = info[15][2:]

			// Checks if index is off and fixes it using a search
			if pack.InstalledSize == "None" {
				index := -1
				spl := strings.Split(string(siOut), "\n")
				for i, s := range spl {
					if strings.Contains(s, "Download") {
						index = len(spl) - i - 2
						break
					}
				}
				pack.InstalledSize = info[len(info)-index+1][2:]
				pack.DownloadSize = info[len(info)-index][2:]
			}

		}

		// Print
		if print {
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

		packs = append(packs, pack)
	}

	return packs, nil

}
