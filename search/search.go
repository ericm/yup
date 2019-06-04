package search

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ericm/yup/output"
	"github.com/mikkeloscar/aur"
)

func setColor(repo *string) {
	// Set colour for repo
	switch *repo {
	case "core":
		// Purple
		*repo = "\033[95mcore\033[0m"
		break
	case "extra":
		// Green
		*repo = "\033[32mextra\033[0m"
		break
	case "community":
		// Cyan
		*repo = "\033[36mcommunity\033[0m"
		break
	case "multilib":
		// Yellow
		*repo = "\033[33mmultilib\033[0m"
		break
	}
}

// Aur returns []Package parsed from the AUR
func Aur(query string, print bool, installed bool) ([]output.Package, error) {
	// Hardcoded query limit
	limit := 100
	i := 0

	// Generate query
	queryS := strings.Split(query, " ")
	aurPacksB := []aur.Pkg{}
	aurPacks := []aur.Pkg{}

	// Search the AUR
	for _, q := range queryS {
		aurPack, err := aur.Search(q)
		if err != nil {
			return []output.Package{}, err
		}
		aurPacksB = append(aurPacksB, aurPack...)
	}

	// Filter aurPacksB (before) to aurPacks
	for _, pack := range aurPacksB {
		matched := true

		for _, q := range queryS {
			if !((strings.Contains(pack.Name, q) || strings.Contains(pack.Description, q)) && matched) {
				matched = false
			}
		}

		if matched && sort.Search(len(aurPacks), func(i int) bool { return aurPacks[i].Name == pack.Name }) >= len(aurPacks) {
			aurPacks = append(aurPacks, pack)
		}

	}

	packs := []output.Package{}

	for _, pack := range aurPacks {
		newPack := output.Package{
			Aur:         true,
			Name:        pack.Name,
			Repo:        "\033[91maur\033[0m",
			Description: pack.Description,
			Version:     pack.Version,
		}

		// Check if its installed
		ins, errCheck := PacmanQi(newPack.Name)
		if len(ins) > 0 && errCheck == nil {
			newPack.Installed = true
			newPack.InstalledSize = ins[0].InstalledSize
			newPack.InstalledSizeInt = ins[0].InstalledSizeInt
			newPack.DownloadSize = ins[0].DownloadSize
		}

		if print {
			output.PrintPackage(newPack)
		}

		if i < limit {
			packs = append(packs, newPack)
		} else {
			return packs, nil
		}
		i++

	}
	return packs, nil
}

// Pacman returns []Package parsed from pacman
func Pacman(query string, print bool, installed bool) ([]output.Package, error) {
	// Split query
	search := exec.Command("pacman", append([]string{"-Ss"}, strings.Split(query, " ")...)...)
	run, err := search.Output()
	if err != nil {
		return []output.Package{}, nil
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

	packs := []output.Package{}
	for _, pac := range pacOut {
		pack := output.Package{
			Name:        nameRe.FindString(pac)[1:],
			Repo:        repoRe.FindString(pac),
			Version:     strings.Split(versionRe.FindString(pac), " ")[1],
			Installed:   len(installedRe.FindString(pac)) != 0,
			Description: strings.Split(pac, "\n")[1][4:],
		}

		setColor(&pack.Repo)

		if installed {
			query = "="
		}
		if pack.Installed && len(query) > 0 {
			// Add extra install info
			// Get info from pacman -Sii package
			// Add extra install info
			pacmanSi := exec.Command("pacman", "-Sii", pack.Name)
			siOut, err := pacmanSi.Output()
			if err != nil {
				return []output.Package{}, err
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
			output.PrintPackage(pack)
		}

		if installed && pack.Installed {
			packs = append(packs, pack)
		} else if !installed {
			packs = append(packs, pack)
		}

	}

	return packs, nil

}

// PacmanQi parses Installed only from pacman -Qi
func PacmanQi(arg ...string) ([]output.Package, error) {
	out := []output.Package{}

	args := []string{"-Qi"}
	args = append(args, arg...)
	pacmanSi := exec.Command("pacman", args...)
	siOut, err := pacmanSi.Output()
	if err != nil {
		return []output.Package{}, err
	}

	siRe := regexp.MustCompile("(?:\\:)(.+)")

	// Get each pack
	packs := strings.Split(string(siOut), "\n\n")
	for _, pack := range packs {
		parts := siRe.FindAllString(string(pack), -1)
		if len(parts) > 0 {
			// Package it into the object
			newPack := output.Package{
				Name:          parts[0][2:],
				Version:       parts[1][2:],
				Description:   parts[2][2:],
				InstalledSize: parts[len(parts)-7][2:],
				Installed:     true,
			}

			newPack.InstalledSizeInt = ToBytes(newPack.InstalledSize)
			out = append(out, newPack)
		}

	}

	return out, nil
}

type sortPack []output.Package

func (s sortPack) Len() int           { return len(s) }
func (s sortPack) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortPack) Less(i, j int) bool { return s[i].SortValue < s[j].SortValue }

// SortPacks is used to generate the dialogue for yup <query>
func SortPacks(queryS string, packs []output.Package) {
	// Replace query spaces with '-'
	query := strings.ReplaceAll(queryS, " ", "-")
	querySpl := strings.Split(query, "-")[0]

	// Set sort weighting
	for i, pack := range packs {
		// See how close the package name is to the query

		// Check for exact match
		if packs[i].Name == query {
			packs[i].SortValue = 1
			continue
		}

		name := float64(len(pack.Name))
		q := float64(len(query))

		// Check for partial match
		if strings.Contains(pack.Name, query) {
			packs[i].SortValue = 1 / (name / q)
			continue
		}

		// Else one part of the query
		if strings.Contains(pack.Name, querySpl) {
			packs[i].SortValue = 1 / (name / q) / 2
			continue
		}
	}

	sort.Sort(sortPack(packs))

	for i, pack := range packs {
		fmt.Printf("\033[37m\033[1m%d \033[0m", len(packs)-i)
		output.PrintPackage(pack)
	}
}

// ToBytes Turns 1 KiB into 1024
func ToBytes(data string) int {
	valF, err := strconv.ParseFloat(data[:len(data)-4], 32)
	if err != nil {
		return 0
	}
	val := int(valF)
	switch data[len(data)-3:] {
	case "KiB":
		return val * 1000
	case "MiB":
		return val * 1000000
	case "GiB":
		return val * 1000000000
	default:
		return -1
	}
}
