package search

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ericm/goncurses"
	"github.com/ericm/yup/config"
	"github.com/ericm/yup/output"
	"github.com/ericm/yup/sync"
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
			newPack.InstalledVersion = ins[0].Version
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
				Name:             parts[0][2:],
				Version:          parts[1][2:],
				InstalledVersion: parts[1][2:],
				Description:      parts[2][2:],
				InstalledSize:    parts[len(parts)-7][2:],
				Installed:        true,
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

	// Prints using ncurses
	if config.GetConfig().UserFile.Ncurses {
		printncurses(&packs)
	} else {

	Redo:
		for i, pack := range packs {
			fmt.Print("\033[37m\033[1m")
			fmt.Printf("%-5s", fmt.Sprintf("(%d)", len(packs)-i))
			fmt.Print("\033[0m")
			output.PrintPackage(pack, "def")
		}

		packsToInstall := []output.Package{}

		output.PrintL()
		output.Printf("Click on a package above")

		// Read Stdin
		output.PrintIn("Or type packages to install (eg: 1 2 3, 1-3 or ^4)")
		scanner := bufio.NewReader(os.Stdin)
		input, _ := scanner.ReadString('\n')

		inputs := strings.Split((strings.ToLower(strings.TrimSpace(input))), " ")
		seen := map[int]bool{}
		for _, s := range inputs {
			// 1-3
			if strings.Contains(s, "-") {
				continue
			}
			// ^4
			if strings.Contains(s, "^") {
				continue
			}

			if num, err := strconv.Atoi(s); err == nil {
				// Find package from input
				index := len(packs) - num
				// Add to the slice
				if index < len(packs) && index >= 0 && !seen[index] {
					packsToInstall = append(packsToInstall, packs[index])
					seen[index] = true
				}
			}
		}

		// Print packs
		output.Printf("The following packages will be installed:")
		for i, pack := range packsToInstall {
			fmt.Printf("    %-2d \033[1m%s\033[0m %s (%s)\n", i+1, pack.Name, pack.Version, pack.Repo)
		}

		// Ask if they want to redo
		output.PrintIn("Redo selection? (y/N)")
		redo, _ := scanner.ReadString('\n')
		switch strings.ToLower((strings.TrimSpace(redo))) {
		case "y":
			goto Redo
		default:
			break
		}

		// Then, install the packages
		for _, pack := range packsToInstall {
			sync.Sync([]string{pack.Name}, pack.Aur, false)
		}
	}

}

// Prints ncurses
func printncurses(packs *[]output.Package) {
	// Setup ncurses
	stdscr, err := goncurses.Init()
	if err != nil {
		output.PrintErr("%s", err)
	}
	defer goncurses.End()

	goncurses.Cursor(0)
	goncurses.Echo(false)
	goncurses.Raw(true)

	// Init the ncurses colours
	goncurses.StartColor()
	goncurses.InitPair(1, goncurses.C_RED, goncurses.C_BLACK)
	goncurses.InitPair(11, goncurses.C_RED, goncurses.C_WHITE)
	goncurses.InitPair(2, goncurses.C_CYAN, goncurses.C_BLACK)
	goncurses.InitPair(12, goncurses.C_CYAN, goncurses.C_WHITE)
	goncurses.InitPair(3, goncurses.C_YELLOW, goncurses.C_BLACK)
	goncurses.InitPair(13, goncurses.C_YELLOW, goncurses.C_WHITE)
	goncurses.InitPair(4, goncurses.C_GREEN, goncurses.C_BLACK)
	goncurses.InitPair(14, goncurses.C_GREEN, goncurses.C_WHITE)
	goncurses.InitPair(5, goncurses.C_MAGENTA, goncurses.C_BLACK)
	goncurses.InitPair(15, goncurses.C_MAGENTA, goncurses.C_WHITE)
	goncurses.InitPair(6, goncurses.C_WHITE, goncurses.C_BLACK)
	goncurses.InitPair(16, goncurses.C_WHITE, goncurses.C_WHITE)
	// Selected
	goncurses.InitPair(7, goncurses.C_BLACK, goncurses.C_WHITE)

	// Menu
	goncurses.InitPair(8, goncurses.C_BLUE, goncurses.C_BLACK)

	// Initial print
	selected := 1
	printPacks(stdscr, packs, selected)
	printBar(stdscr)

	stdscr.Refresh()

	// Event loop
	var ch goncurses.Key
	for ch != 'q' {
		ch = stdscr.GetChar()
		update := false
		switch ch {
		case 65, 'w':
			// Scroll forward
			if selected < len(*packs)-1 {
				selected += 1
				update = true
			}
		case 66, 's':
			// Scroll backward
			if selected > 1 {
				selected -= 1
				update = true
			}
		}

		if update {
			printPacks(stdscr, packs, selected)
			printBar(stdscr)
			stdscr.ClearToEOL()
			stdscr.Refresh()
		}
	}
}

func printBar(stdscr *goncurses.Window) {
	my, mx := stdscr.MaxYX()

	// Print line
	stdscr.ColorOn(8)
	line := ""
	for _i := 0; _i < mx; _i++ {
		line += "="
	}
	stdscr.MovePrintf(my-3, 0, "%s", line)
	stdscr.ColorOff(8)

	// Print Input
	stdscr.ColorOn(5)
	stdscr.MovePrint(my-2, 0, "==>")
	stdscr.ColorOff(5)
	stdscr.MovePrint(my-2, 4, "Click on a package above, use the arrow keys and enter")

	stdscr.ColorOn(4)
	stdscr.MovePrint(my-1, 0, "==> Or type packages to install (eg: 1 2 3, 1-3 or ^4):")
	stdscr.ColorOff(4)
}

func printPacks(stdscr *goncurses.Window, packs *[]output.Package, selected int) {
	my, mx := stdscr.MaxYX()
	for i, item := range *packs {
		sel := len(*packs)-i == selected
		y := my - (2 * (len(*packs) - i)) - 3

		// Number
		if sel {
			stdscr.ColorOn(7)
		}
		stdscr.AttrOn(goncurses.A_BOLD)
		stdscr.MovePrintf(y, 0, "(%d)", len(*packs)-i)
		stdscr.AttrOff(goncurses.A_BOLD)
		if sel {
			stdscr.ColorOff(7)
		}

		cur := 5

		// Repo
		switch item.Repo {
		case "\033[91maur\033[0m":
			cur += 3
			stdscr.ColorOn(1)
			stdscr.MovePrint(y, 5, "aur")
			stdscr.ColorOff(1)
		case "\033[95mcore\033[0m":
			cur += 4
			stdscr.ColorOn(5)
			stdscr.MovePrint(y, 5, "core")
			stdscr.ColorOff(5)
		case "\033[32mextra\033[0m":
			cur += 5
			stdscr.ColorOn(4)
			stdscr.MovePrint(y, 5, "extra")
			stdscr.ColorOff(4)
		case "\033[36mcommunity\033[0m":
			cur += 9
			stdscr.ColorOn(2)
			stdscr.MovePrint(y, 5, "community")
			stdscr.ColorOff(2)
		case "\033[33mmultilib\033[0m":
			cur += 8
			stdscr.ColorOn(3)
			stdscr.MovePrint(y, 5, "multilib")
			stdscr.ColorOff(3)
		}

		// Slash
		stdscr.AttrOn(goncurses.A_DIM)
		stdscr.MovePrint(y, cur, "/")
		cur += 1
		stdscr.AttrOff(goncurses.A_DIM)

		// Name
		stdscr.AttrOn(goncurses.A_BOLD)
		stdscr.MovePrint(y, cur, item.Name)
		cur += len(item.Name) + 1
		stdscr.AttrOff(goncurses.A_BOLD)

		// Version
		stdscr.MovePrint(y, cur, item.Version)
		cur += len(item.Version) + 1

		// Installed
		if item.Installed {
			stdscr.MovePrint(y, cur, "(")
			cur += 1
			stdscr.AttrOn(goncurses.A_BOLD)
			stdscr.ColorOn(5)
			stdscr.MovePrint(y, cur, "INSTALLED")
			cur += 9
			if item.InstalledVersion != item.Version {
				// Outdated
				cur += 1
				stdscr.MovePrint(y, cur, "OUTDATED")
				stdscr.ColorOff(5)
				stdscr.AttrOff(goncurses.A_BOLD)
				cur += 9
				stdscr.MovePrint(y, cur, item.InstalledVersion)
				cur += len(item.InstalledVersion)
			}
			stdscr.ColorOff(5)
			stdscr.AttrOff(goncurses.A_BOLD)
			stdscr.MovePrint(y, cur, ")")
			// Size
			cur += 2
			stdscr.MovePrintf(y, cur, "Size: (Dl: %s | Ins: %s)", item.DownloadSize, item.InstalledSize)
		}

		// Description
		desc := item.Description
		if len(desc) > mx-7 {
			desc = desc[:(mx-9)] + ".."
		}
		stdscr.MovePrintf(y+1, 5, "- %s", desc)
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
