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

	"time"

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
	// Query limit
	limit := config.GetConfig().UserFile.AurLimit

	// Generate query
	queryS := strings.Split(query, " ")
	aurPacks := []aur.Pkg{}

	// Search the AUR
	aurPackIn, err := aur.Search(query)
	if err != nil {
		return []output.Package{}, err
	}

	secondaryAur, err := aur.Search(strings.ReplaceAll(query, " ", "-"))
	aurPackIn = append(aurPackIn, secondaryAur...)

	seen := map[string]bool{}
	// Filter aurPacksB (before) to aurPacks
	for _, pack := range aurPackIn {
		matched := true

		for _, q := range queryS {
			if !((strings.Contains(pack.Name, q) || strings.Contains(pack.Description, q)) && matched) {
				matched = false
			}
		}

		if matched && sort.Search(len(aurPacks), func(i int) bool { return aurPacks[i].Name == pack.Name }) >= len(aurPacks) {
			if !seen[pack.Name] {
				seen[pack.Name] = true
				aurPacks = append(aurPacks, pack)
			}
		}

	}

	packs := []output.Package{}

	for i, pack := range aurPacks {
		newPack := output.Package{
			Aur:         true,
			Name:        pack.Name,
			Repo:        "\033[91maur\033[0m",
			Description: pack.Description,
			Version:     pack.Version,
			OutOfDate:   pack.OutOfDate,
			Upstream:    pack.URL,
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

	// Sii regex
	sizeRe := regexp.MustCompile("Installed Size\\s*?:\\s*?(\\d*.\\d*\\s\\w+)")
	urlRe := regexp.MustCompile("URL\\s*?:\\s*?(\\s\\S+)")
	insVerRe := regexp.MustCompile("Version\\s*?:\\s*(\\s\\S+)")

	packs := []output.Package{}
	for i, pac := range pacOut {
		if i > config.GetConfig().UserFile.PacmanLimit-1 {
			// Query limit
			break
		}
		pack := output.Package{
			Name:        nameRe.FindString(pac)[1:],
			Repo:        repoRe.FindString(pac),
			Version:     strings.Split(versionRe.FindString(pac), " ")[1],
			Installed:   len(installedRe.FindString(strings.Split(pac, "\n")[0])) != 0,
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
			pacmanSi := exec.Command("pacman", "-Qi", pack.Name)
			siOut, err := pacmanSi.Output()
			if err != nil {
				return []output.Package{}, err
			}

			// Sets the other vals
			info := siRe.FindAllString(string(siOut), -1)
			size := sizeRe.FindAllString(string(siOut), -1)
			insVer := insVerRe.FindAllString(string(siOut), -1)
			upstream := urlRe.FindStringSubmatch(string(siOut))

			if len(insVer) > 0 && len(insVer[0]) > 18 {
				pack.InstalledVersion = insVer[0][18:]
			}
			if len(size) > 0 && len(size[0]) > 18 {
				pack.InstalledSize = size[0][18:]
			}
			if len(upstream) > 0 && len(upstream[0]) > 18 {
				pack.Upstream = urlRe.FindStringSubmatch(string(siOut))[0][18:]
			}
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

func PacmanGroups(term string) ([]output.Package, error) {
	term = strings.ToLower(term)
	pac := exec.Command("pacman", "-Sg")
	out, err := pac.Output()
	if err != nil {
		return nil, err
	}
	outS := string(out)
	packs := []output.Package{}
	for _, s := range strings.Split(outS, "\n") {
		packs = append(packs, output.Package{
			Name:        s,
			Repo:        "\033[94mgroup\033[0m",
			Description: fmt.Sprintf("%s package group", s),
		})
	}

	outPacks := []output.Package{}

	for _, pack := range packs {
		if strings.Contains(pack.Name, term) {
			outPacks = append(outPacks, pack)
		}
	}

	return outPacks, nil
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
	sizeRe := regexp.MustCompile("Installed Size\\s*?:\\s*?(\\d*.\\d*\\s\\w+)")

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
				InstalledSize:    sizeRe.FindAllString(string(pack), -1)[0][18:],
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

	if mode := config.GetConfig().UserFile.SortMode; mode == "closest" {
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
	}

	if len(packs) == 0 {
		output.PrintErr("No results found")
		return
	}

	if config.GetConfig().UserFile.SortMode != "none" {
		sort.Sort(sortPack(packs))
	}

	scanner := bufio.NewReader(os.Stdin)
	packsToInstall := []output.Package{}
Redo:
	// Prints using ncurses
	conf := config.GetConfig()
	if !conf.Ncurses || (conf.UserFile.Ncurses && !conf.Ncurses) {
		if newPacks, check := printncurses(&packs); check {
			packsToInstall = newPacks
		} else if newPacks != nil {
			// Remove
			for _, pac := range newPacks {
				rem := exec.Command("sudo", "pacman", "-R", pac.Name)
				output.SetStd(rem)
				if err := rem.Run(); err != nil {
					output.PrintErr("%s", err)
				}
			}
			os.Exit(1)
		} else {
			os.Exit(1)
		}
	} else {

		for i, pack := range packs {
			fmt.Print("\033[37m\033[1m")
			fmt.Printf("%-5s", fmt.Sprintf("(%d)", len(packs)-i))
			fmt.Print("\033[0m")
			output.PrintPackage(pack, "def")
		}

		output.PrintL()
		output.Printf("Click on a package above")

		// Read Stdin
		output.PrintIn("Or type packages to install (eg: 1 2 3, 1-3 or ^4)")
		input, _ := scanner.ReadString('\n')

		inputs := strings.Split((strings.ToLower(strings.TrimSpace(input))), " ")
		seen := map[int]bool{}
		for _, s := range inputs {
			// 1-3
			if strings.Contains(s, "-") {
				if spl := strings.Split(s, "-"); len(spl) == 2 {
					// Get int vals for range
					firstT, errF := strconv.Atoi(spl[0])
					secondT, errS := strconv.Atoi(spl[1])
					if errF == nil && errS == nil {
						// Convert range from visual representation
						first := len(packs) - firstT
						second := len(packs) - secondT
						// Filter
						for i := second; i <= first; i++ {
							packsToInstall = append(packsToInstall, packs[i])
						}
					}
				}
				continue
			}
			// ^4
			if strings.Contains(s, "^") {
				if num, err := strconv.Atoi(s[1:]); err == nil {
					// Filter for the number
					for i, pack := range packs {
						ind := len(packs) - i
						if ind == num {
							continue
						}
						packsToInstall = append(packsToInstall, pack)
					}
				}
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
		packs = packsToInstall

	}
	// Print packs
	output.Printf("The following packages will be installed:")
	for i, pack := range packsToInstall {
		fmt.Printf("    %-2d \033[1m%s\033[0m %s (%s)\n", i+1, pack.Name, pack.Version, pack.Repo)
	}

	if conf.UserFile.AskRedo {
		// Ask if they want to redo
		output.PrintIn("Redo selection? (y/N)")
		redo, _ := scanner.ReadString('\n')
		switch strings.ToLower((strings.TrimSpace(redo))) {
		case "y":
			goto Redo
		default:
			break
		}
	}

	// Then, install the packages
	for _, pack := range packsToInstall {
		sync.Sync([]string{pack.Name}, pack.Aur, false)
	}

}

func getDims() (string, string) {
	var (
		prevMy string
		prevMx string
	)
	if prevMyB, err := exec.Command("tput", "lines").Output(); err == nil {
		prevMy = string(prevMyB)
	}
	if prevMxB, err := exec.Command("tput", "cols").Output(); err == nil {
		prevMx = string(prevMxB)
	}

	return prevMx, prevMy
}

// Prints ncurses
func printncurses(packs *[]output.Package) ([]output.Package, bool) {
	selected := 1
	checked := map[int]bool{}

Resize:
	stdscr, err := goncurses.Init()
	if err != nil {
		output.PrintErr("%s", err)
	}
	stdscr.Keypad(true)
	defer goncurses.End()

	goncurses.Cursor(0)
	goncurses.Echo(false)
	goncurses.Raw(true)

	goncurses.MouseMask(goncurses.M_ALL, nil) // temporarily enable all mouse clicks
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
	goncurses.InitPair(9, goncurses.C_BLACK, goncurses.C_MAGENTA)

	// Menu
	goncurses.InitPair(8, goncurses.C_BLUE, goncurses.C_BLACK)

	goncurses.InitPair(10, goncurses.C_BLACK, goncurses.C_WHITE)

	// Event loop
	// Setup ncurses
	var ch goncurses.Key
	var offset int
	var timeout bool
	var newSel int
	var notSel bool
	var toSel int
	prevMy := ""
	prevMx := ""

	// Initial print
	printPacks(stdscr, packs, selected, checked)
	printBar(stdscr, 0, 0, false)
	printhelp(stdscr)

	stdscr.Refresh()

	for ch != 'q' && ch != 27 {
		update := false
		if !timeout {
		Sw:
			switch goncurses.Key(ch) {
			case goncurses.KEY_UP, goncurses.KEY_SF, 'w':
				// Scroll forward
				if selected < len(*packs) {
					selected += 1
					update = true
				}
			case goncurses.KEY_DOWN, goncurses.KEY_SR, 's':
				// Scroll backward
				if selected > 1 {
					selected -= 1
					update = true
				}
			case goncurses.KEY_MOUSE:
				if ms := goncurses.GetMouse(); ms != nil {
					if ms.State == goncurses.M_B1_CLICKED {
						clicked := -1
						my, _ := stdscr.MaxYX()
						clicked = getactive(ms.Y, my, offset, selected, packs)
						if clicked != -1 {
							checked[clicked] = !checked[clicked]
						}
					} else if ms.State == goncurses.M_B4_PRESSED {
						// Scroll up
						ch = goncurses.KEY_SF
						goto Sw
					} else if ms.State == goncurses.M_B5_PRESSED {
						// Scroll down
						ch = goncurses.KEY_SR
						goto Sw
					}
					update = true
				}

			case '\n', ' ':
				if notSel {
					for i := 0; i <= len(*packs); i++ {
						if i == newSel {
							continue
						}
						checked[i] = true
					}
					update = true
					notSel = false
					toSel = 0
					newSel = 0
					break
				}

				if newSel != 0 {
					if toSel != 0 {
						if toSel > newSel {
							for i := newSel; i <= toSel; i++ {
								checked[i] = true
								update = true
							}
							selected = toSel
						} else if toSel < newSel {
							for i := toSel; i <= newSel; i++ {
								checked[i] = true
								update = true
							}
							selected = newSel
						}
					} else {
						num := newSel
						if num > 0 && num < len(*packs) {
							checked[num] = !checked[num]
							selected = num
							update = true
						}
					}

				} else {
					checked[selected] = !checked[selected]
					update = true
				}
				newSel = 0
				toSel = 0

			case 'u':
				cm := exec.Command("xdg-open", (*packs)[len(*packs)-selected].Upstream)
				cm.Run()

			case 'i', 'z', 'r':
				// Filter packs
				newPack := []output.Package{}
				for i, pack := range *packs {
					if checked[len(*packs)-i] {
						newPack = append(newPack, pack)
					}
				}

				// Check if none selected
				if len(newPack) == 0 {
					newPack = append(newPack, (*packs)[len(*packs)-selected])
				}

				if ch == 'r' {
					return newPack, false
				} else {
					return newPack, true
				}

			case '-':
				if newSel != 0 && !notSel {
					toSel = newSel
					newSel = 0
					update = true
				}

			case '^':
				if !notSel {
					notSel = true
					update = true
				}
			case 'f':
				goncurses.End()
				goto Resize

			default:
				if num, err := strconv.Atoi(string(ch)); err == nil {
					if newSel != 0 {
						newSel = newSel*10 + num
					} else {
						newSel = num
					}
					update = true
				}
			}

		}

		// Get prev size
		if prevMx == "" || prevMy == "" {
			prevMx, prevMy = getDims()
		} else if mmx, mmy := getDims(); prevMx != mmx || prevMy != mmy {
			goncurses.End()
			goto Resize
		}
		// Mouse timeout
		timeout = true
		go func(timeout *bool) {
			time.Sleep(30 * time.Millisecond)
			*timeout = false
		}(&timeout)
		if update {
			stdscr.Clear()
			offset = printPacks(stdscr, packs, selected, checked)
			printBar(stdscr, newSel, toSel, notSel)
			printhelp(stdscr)
		}
		ch = stdscr.GetChar()

	}
	return nil, false
}

func getactive(y, my, offset, selected int, packs *[]output.Package) int {
	if y >= my-3 {
		return -1
	}

	for i := range *packs {
		ind := len(*packs) - i - 3
		iy := my - (2 * (len(*packs) - i)) + offset + 3

		if y == iy || y == iy+1 {
			return ind
		}

	}
	return -1
}

func printBar(stdscr *goncurses.Window, newSel, toSel int, notSel bool) int {
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

	// Print user input
	cur := 56
	if notSel {
		stdscr.MovePrint(my-1, cur, "^")
		cur++
	}
	if toSel != 0 {
		ts := strconv.Itoa(toSel)
		stdscr.MovePrintf(my-1, cur, ts+"-")
		cur += len(ts) + 1
	}
	if newSel != 0 {
		stdscr.MovePrint(my-1, cur, newSel)
	}

	return my
}

func printPacks(stdscr *goncurses.Window, packs *[]output.Package, selected int, checked map[int]bool) int {
	my, mx := stdscr.MaxYX()
	// Calculate offset up
	offset := 0
	if selected*2 > my-5 {
		offset = selected*2 - my + 5
	}
	for i, item := range *packs {
		ind := len(*packs) - i
		sel := ind == selected
		check := checked[ind]
		y := my - (2 * (len(*packs) - i)) - 3 + offset

		if y > my-4 {
			continue
		}
		// Number
		if sel {
			stdscr.ColorOn(7)
		} else if check {
			stdscr.ColorOn(9)
		}
		stdscr.AttrOn(goncurses.A_BOLD)
		if check {
			stdscr.MovePrintf(y, 0, "%-5s", fmt.Sprintf("(%d)", len(*packs)-i))
		} else {
			stdscr.MovePrintf(y, 0, "(%d)", len(*packs)-i)
		}
		stdscr.AttrOff(goncurses.A_BOLD)
		if sel {
			stdscr.ColorOff(7)
		} else if check {
			stdscr.ColorOff(9)
		}

		cur := 5

		// Repo
		switch item.Repo {
		case "\033[91maur\033[0m":
			cur += 3
			if check {
				stdscr.ColorOn(9)
			} else {
				stdscr.ColorOn(1)
			}
			stdscr.MovePrint(y, 5, "aur")
			if check {
				stdscr.ColorOff(9)
			} else {
				stdscr.ColorOff(1)
			}
		case "\033[95mcore\033[0m":
			cur += 4
			if check {
				stdscr.ColorOn(9)
			} else {
				stdscr.ColorOn(5)
			}
			stdscr.MovePrint(y, 5, "core")
			if check {
				stdscr.ColorOff(9)
			} else {
				stdscr.ColorOff(5)
			}
		case "\033[32mextra\033[0m":
			cur += 5
			if check {
				stdscr.ColorOn(9)
			} else {
				stdscr.ColorOn(4)
			}
			stdscr.MovePrint(y, 5, "extra")
			if check {
				stdscr.ColorOff(9)
			} else {
				stdscr.ColorOff(4)
			}
		case "\033[36mcommunity\033[0m":
			cur += 9
			if check {
				stdscr.ColorOn(9)
			} else {
				stdscr.ColorOn(2)
			}
			stdscr.MovePrint(y, 5, "community")
			if check {
				stdscr.ColorOff(9)
			} else {
				stdscr.ColorOff(2)
			}
		case "\033[33mmultilib\033[0m":
			cur += 8
			if check {
				stdscr.ColorOn(9)
			} else {
				stdscr.ColorOn(3)
			}
			stdscr.MovePrint(y, 5, "multilib")
			if check {
				stdscr.ColorOff(9)
			} else {
				stdscr.ColorOff(3)
			}
		case "\033[94mgroup\033[0m":
			cur += 5
			if check {
				stdscr.ColorOn(9)
			} else {
				stdscr.ColorOn(8)
			}
			stdscr.MovePrint(y, 5, "group")
			if check {
				stdscr.ColorOff(9)
			} else {
				stdscr.ColorOff(8)
			}
		default:
			cur += len(item.Repo)
			if check {
				stdscr.ColorOn(9)
			}
			stdscr.MovePrint(y, 5, item.Repo)
			if check {
				stdscr.ColorOff(9)
			}
		}

		// Slash
		if check {
			stdscr.ColorOn(9)
		}
		stdscr.AttrOn(goncurses.A_DIM)
		stdscr.MovePrint(y, cur, "/")
		cur += 1
		stdscr.AttrOff(goncurses.A_DIM)
		if check {
			stdscr.ColorOff(9)
		}

		if check {
			stdscr.ColorOn(9)
		}
		// Name
		stdscr.AttrOn(goncurses.A_BOLD)
		stdscr.MovePrint(y, cur, item.Name)
		cur += len(item.Name) + 1
		stdscr.AttrOff(goncurses.A_BOLD)

		if check {
			stdscr.ColorOff(9)
		}

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
			stdscr.MovePrintf(y, cur, "Install Size: %s", item.InstalledSize)
			cur += 14 + len(item.InstalledSize)
		}

		// Out of date
		if item.OutOfDate != 0 {
			stdscr.AttrOn(goncurses.A_BOLD)
			stdscr.ColorOn(1)
			stdscr.MovePrintf(y, cur, "(OUT OF DATE)")
			stdscr.ColorOff(1)
			stdscr.AttrOff(goncurses.A_BOLD)
			cur += 13
		}

		// Description
		desc := item.Description
		if len(desc) > mx-7 {
			desc = desc[:(mx-9)] + ".."
		}
		stdscr.MovePrintf(y+1, 5, "- %s", desc)
	}

	return offset
}

// Help
func printhelp(stdscr *goncurses.Window) {
	_, mx := stdscr.MaxYX()
	stdscr.ColorOn(10)
	stdscr.MovePrintf(0, mx-15, " %-14s", "Enter: Select")
	stdscr.MovePrintf(1, mx-15, " %-14s", "I/Z: Install")
	stdscr.MovePrintf(2, mx-15, " %-14s", "R: Remove")
	stdscr.MovePrintf(3, mx-15, " %-14s", "U: Upstream")
	stdscr.MovePrintf(4, mx-15, " %-14s", "Q: Quit")
	stdscr.ColorOff(10)
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
