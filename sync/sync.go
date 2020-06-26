package sync

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Morganamilo/go-srcinfo"
	"github.com/ericm/yup/output"
	"github.com/mikkeloscar/aur"

	"fmt"

	"strconv"

	"github.com/ericm/yup/config"
)

// Download wrapper for io.Reader
type Download struct {
	io.Reader
	total int64
	count int
}

// PkgBuild represents a package that has been read from the AUR
type PkgBuild struct {
	file        string
	dir         string
	name        string
	version     string
	depends     []string
	makeDepends []string
	optDepends  []string
	update      bool
	pacman      bool
}

// Sync from the AUR first, then other configured repos.
//
// This checks each package param individually
func Sync(packages []string, isAur bool, silent bool) error {
	if len(packages) > 0 && len(packages[0]) == 0 {
		return fmt.Errorf("No targets specified (use -h for help)")
	}

	// Create channels for goroutines
	// Step 1: Check AUR
	errChannel := make(chan error, len(packages))
	buildChannel := make(chan *PkgBuild, len(packages))

	pacmanArgs := []string{}

	for _, p := range packages {
		// Multithreaded downloads
		go func(p string) {
			// If designated, install from pacman
			if !isAur {
				buildChannel <- nil
				errChannel <- nil
				pacmanArgs = append(pacmanArgs, p)
			}
			repo, err := aur.Info([]string{p})
			if err != nil {
				errChannel <- err
			} else {
				if len(repo) > 0 {
					aurDload("https://aur.archlinux.org/"+repo[0].PackageBase+".git", errChannel, buildChannel, repo[0].PackageBase, repo[0].Version, repo[0].Depends, repo[0].MakeDepends, repo[0].OptDepends)
				} else {
					errChannel <- output.Errorf("Didn't find an \033[1mAUR\033[0m package for \033[1m\033[32m%s\033[39m\033[0m, searching other repos\n", p)
					buildChannel <- nil
					pacmanArgs = append(pacmanArgs, p)
				}
			}

		}(p)
	}

	for _i := 0; _i < len(packages)*2; _i++ {
		// Check for both error and build Channels
		select {
		case err := <-errChannel:
			if err != nil && !silent {
				fmt.Print(err)
			}
		case pkg := <-buildChannel:
			if pkg != nil {
				// Install the package
				if err := pkg.Install(silent, false); err != nil {
					return err
				}
			}
		}

	}

	// Now check pacman for unresolved args in pacmanArgs
	if len(pacmanArgs) > 0 {
		sync := pacmanSync(pacmanArgs, false, false)
		for _, s := range sync {
			if s != nil {
				return s
			}
		}
	}

	return nil
}

// ParseNumbers filters according to user input
func ParseNumbers(input string, packs *[]PkgBuild) {
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
					first := len(*packs) - firstT
					second := len(*packs) - secondT
					// Filter
					for i := second; i <= first; i++ {
						seen[i] = true
					}
				}
			}
			continue
		}
		// ^4
		if strings.Contains(s, "^") {
			if num, err := strconv.Atoi(s[1:]); err == nil {
				// Filter for the number
				for i := range *packs {
					ind := len(*packs) - i
					if ind == num {
						continue
					}
					seen[ind] = true
				}
			}
			continue
		}

		if num, err := strconv.Atoi(s); err == nil {
			// Find package from input
			index := len(*packs) - num
			// Add to the slice
			if index < len(*packs) && index >= 0 && !seen[index] {
				seen[index] = true
			}
		}
	}

	newPacks := *packs

	for i := range *packs {
		if seen[i] {
			newPacks = append(newPacks[:i], newPacks[i+1:]...)
		}
	}

	*packs = newPacks

}

// ParseNumbersStr filters according to user input
func ParseNumbersStr(input string, packs *[]string) {
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
					first := len(*packs) - firstT
					second := len(*packs) - secondT
					// Filter
					for i := second; i <= first; i++ {
						seen[i] = true
					}
				}
			}
			continue
		}
		// ^4
		if strings.Contains(s, "^") {
			if num, err := strconv.Atoi(s[1:]); err == nil {
				// Filter for the number
				for i := range *packs {
					ind := len(*packs) - i
					if ind == num {
						continue
					}
					seen[ind] = true
				}
			}
			continue
		}

		if num, err := strconv.Atoi(s); err == nil {
			// Find package from input
			index := len(*packs) - num
			// Add to the slice
			if index < len(*packs) && index >= 0 && !seen[index] {
				seen[index] = true
			}
		}
	}

	newPacks := *packs

	for i := range *packs {
		if seen[i] {
			newPacks = append(newPacks[:i], newPacks[i+1:]...)
		}
	}

	*packs = newPacks

}

// Install the pkgBuild
// assuming repo is now cloned or fetched
func (pkg *PkgBuild) Install(silent, isDep bool) error {
	output.Printf("Installing \033[1m\033[32m%s\033[39m\033[2m %s\033[0m from the AUR", pkg.name, pkg.version)

	// Install from the AUR
	os.Chdir(filepath.Join(pkg.dir, pkg.name))

	scanner := bufio.NewReader(os.Stdin)
	if pkg.update {
		merge := exec.Command("git", "merge", "origin/master")
		merge.Run()
	}
	if !silent && !isDep {
		// Print PkgBuild by default
		conf := config.GetConfig().UserFile
		if conf.PrintPkg {
			output.Printf("PKGBUILD:")
			catPkg := exec.Command("cat", "PKGBUILD")
			output.SetStd(catPkg)
			catPkg.Run()
			fmt.Print("\n")
		}
		if conf.AskPkg {
			i := 0
			output.PrintIn("\033[1m\033[4mV\033[0m\033[92miew, see \033[1m\033[4mD\033[0m\033[92miffs or \033[1m\033[4mE\033[0m\033[92mdit the PKGBUILD? (\033[1m\033[4mA\033[0m\033[92mll or \033[1m\033[4mN\033[0m\033[92mone)")
			out, _ := scanner.ReadString('\n')

			cmds := []*exec.Cmd{}

			// Handle 'a'
			out = strings.TrimSpace(strings.ToLower(out))
			if strings.Contains(out, "a") {
				out = "vde"
			}

		Pkgbuild:
			if i < len(out) {
				switch out[i : i+1] {
				case "v":
					// View
					cmds = append(cmds, exec.Command("cat", "PKGBUILD"))
					i++
					goto Pkgbuild

				case "d":
					// Diffs
					// Get number of commits
					numc := exec.Command("git", "rev-list", "--count", "master")
					nn, _ := numc.Output()
					if string(nn) != strconv.Itoa(1) {
						// If number isnt 1
						var diff *exec.Cmd
						diff = exec.Command("git", "diff", "@~..@")

						cmds = append(cmds, diff)
					}

					i++
					goto Pkgbuild

				case "e":
					// Check for EDITOR
					editor := os.Getenv("EDITOR")
					if len(editor) == 0 {
						// Ask for editor
						output.PrintIn("No EDITOR environment variable set. Enter editor")
						newEditor, _ := scanner.ReadString('\n')
						editor = newEditor[:len(newEditor)-1]
					}

					cmds = append(cmds, exec.Command(editor, "PKGBUILD"))
					i++
					goto Pkgbuild

				case "n":
				case "\n":
					break

				default:
					i++
					goto Pkgbuild
				}
			}

			// Exectue commands
			for _, cmd := range cmds {
				output.SetStd(cmd)
				cmd.Run()
				output.PrintIn("Continue?")
				n, _ := scanner.ReadString('\n')
				if strings.ToLower(n[:1]) == "n" {
					return nil
				}
			}
		}
	}

	// Make / Install the package
	pkg.dir = filepath.Join(pkg.dir, pkg.name)

	os.Chdir(pkg.dir)

	info, err := srcinfo.ParseFile(".SRCINFO")
	if err != nil {
		return err
	}

	if !isDep {
		tempDeps := []string{}
		for _, d := range info.Depends {
			tempDeps = append(tempDeps, d.Value)
		}
		tempMakeDeps := []string{}
		for _, d := range info.MakeDepends {
			tempMakeDeps = append(tempMakeDeps, d.Value)
		}

		// Redefine deps
		pkg.depends, pkg.makeDepends = tempDeps, tempMakeDeps

		remMakes := false
		// Check for dependencies
		output.Printf("Checking for dependencies")
		deps, makeDeps, optDeps, err := pkg.depCheck()
		if err != nil {
			return err
		}

		if len(deps) > 0 {
			output.Printf("Found uninstalled Dependencies:")
			fmt.Print("    ")
			for i, dep := range deps {
				fmt.Printf("\033[1m%d\033[0m %s  ", i+1, dep.name)
			}
			fmt.Print("\n")
			if !silent {
				output.PrintIn("Numbers of packages not to install? (eg: 1 2 3, 1-3 or ^4)")
				depRem, _ := scanner.ReadString('\n')

				// Parse input
				ParseNumbers(depRem, &deps)
			}
		}

		if len(makeDeps) > 0 {
			output.Printf("Found uninstalled Make Dependencies:")
			fmt.Print("    ")
			for i, dep := range makeDeps {
				fmt.Printf("\033[1m%d\033[0m %s  ", i+1, dep.name)
			}
			fmt.Print("\n")

			if !silent {
				// Not to install
				output.PrintIn("Numbers of packages not to install? (eg: 1 2 3, 1-3 or ^4)")

				depNum, _ := scanner.ReadString('\n')
				ParseNumbers(depNum, &makeDeps)

				output.PrintIn("Remove Make Dependencies after install? (y/N)")

				rem, _ := scanner.ReadString('\n')
				switch strings.TrimSpace(strings.ToLower(rem[:1])) {
				case "y":
					remMakes = true
					break
				}
			}
		}
		if len(optDeps) > 0 {
			output.Printf("Found uninstalled Optional Dependencies:")
			fmt.Print("    ")
			for i, dep := range optDeps {
				fmt.Printf("\033[1m%d\033[0m %s  ", i+1, dep.name)
			}
			fmt.Print("\n")
			if !silent {
				output.PrintIn("Numbers of packages TO install? (eg: 1 2 3, 1-3 or ^4)")
				depRem, _ := scanner.ReadString('\n')

				// Parse input
				temp := make([]PkgBuild, len(optDeps))
				copy(temp, optDeps)
				ParseNumbers(depRem, &temp)
				for len(temp) > 0 {
					curr := temp[0]
					if len(temp) == 1 {
						temp = []PkgBuild{}
					} else {
						temp = temp[1:]
					}
					for i, dep := range optDeps {
						if dep.name == curr.name {
							optDeps = append(optDeps[:i], optDeps[i+1:]...)
							break
						}
					}
				}
			}
		}

		// Gather packages
		aurInstall := []PkgBuild{}
		pacInstall := []string{}

		if len(deps) > 0 {
			// Install deps packages
			for i := len(deps) - 1; i >= 0; i-- {
				dep := deps[i]
				fmt.Println(dep.name)
				if dep.pacman {
					// Install from pacman
					pacInstall = append(pacInstall, dep.name)
				} else {
					// Install using Install in silent mode
					aurInstall = append(aurInstall, dep)
				}
			}
		}

		if len(makeDeps) > 0 {
			// Install makeDeps packages
			for i := len(makeDeps) - 1; i >= 0; i-- {
				dep := makeDeps[i]
				if dep.pacman {
					// Install from pacman
					pacInstall = append(pacInstall, dep.name)
				} else {
					// Install using Install in silent mode
					aurInstall = append(aurInstall, dep)
				}
			}

			// At end, remove make packs as necessary
			if remMakes {
				defer func(depM []PkgBuild) {
					output.Printf("Removing Make Dependencies")
					for _, dep := range depM {
						rm := exec.Command("sudo", "pacman", "-R", dep.name)
						output.SetStd(rm)
						if err := rm.Run(); err != nil {
							output.PrintErr("Dep Remove Error: %s", err)
						}
					}
				}(makeDeps)
			}
		}

		if len(optDeps) > 0 {
			// Install deps packages
			for i := len(optDeps) - 1; i >= 0; i-- {
				dep := optDeps[i]
				if dep.pacman {
					// Install from pacman
					pacInstall = append(pacInstall, dep.name)
				} else {
					// Install using Install in silent mode
					aurInstall = append(aurInstall, dep)
				}
			}
		}

		if len(pacInstall) > 0 || len(aurInstall) > 0 {
			output.Printf("Installing Dependencies")
			// Pacman deps
			if err := pacmanSync(pacInstall, true, true); err != nil {
				//output.PrintErr("%s", err)
			}
			// Aur deps
			for _, dep := range aurInstall {
				err := dep.Install(true, true)
				if err != nil {
					output.PrintErr("Dep Install error:")
					return err
				}
				// Set as a dependency
				setDep := exec.Command("sudo", "pacman", "-D", "--asdeps", dep.name)
				if err := setDep.Run(); err != nil {
					return err
				}
			}
		}
	}

	// Change dir back after dep resolution
	os.Chdir(pkg.dir)

	// Get PGP Keys
	for _, key := range info.ValidPGPKeys {
		checkImp := exec.Command("gpg", "--list-keys", "--fingerprint", key)
		if errC := checkImp.Run(); errC != nil {
			// The key isn't imported
			// Ask to import it
			output.PrintIn("Import PGP Key %s? (Y/n)", key)
			check, _ := scanner.ReadString('\n')
			if len(check) > 0 && (strings.ToLower(check)[0] == 'y' || check[0] == '\n') {
				// Import key
				imp := exec.Command("gpg", "--recv-keys", key)
				output.SetStd(imp)
				if err := imp.Run(); err != nil {
					output.PrintErr("%s", err)
				}
			}
		}
	}
	for _, c := range info.Conflicts {
		existsErr := exec.Command("pacman", "-T", pkg.name).Run() // Make sure this isn't an update
		if err := exec.Command("pacman", "-T", c.Value).Run(); err == nil && existsErr != nil {
			// This means that the conflict is installed and we should uninstall
			output.PrintIn(
				"%s and %s are in conflict. Uninstall %s? (y/N)",
				pkg.name,
				c.Value,
				c.Value,
			)
			scan, _ := scanner.ReadString('\n')
			switch strings.TrimSpace(strings.ToLower(scan[:1])) {
			case "y":
				rem := exec.Command("sudo", "pacman", "-R", c.Value)
				output.SetStd(rem)
				if err := rem.Run(); err != nil {
					return err
				}
			}
		}
	}

	// Now, Install the actual package
	cmdMake := exec.Command("makepkg", "-sic", "--noconfirm")
	// Pipe to stdout, etc
	output.SetStd(cmdMake)
	if err := cmdMake.Run(); err != nil {
		return err
	}

	return nil
}

// Download an AUR package to cache
func aurDload(url string, errChannel chan error, buildChannel chan *PkgBuild, name string, version string, depends []string, makeDepends []string, optDepends []string) {
	// TODO: Check in cache
	conf := config.GetConfig()
	dir := filepath.Join(conf.CacheDir, name)

	// Check if git repo is cloned
	update := false
	if os.IsNotExist(os.Chdir(dir)) {
		git := exec.Command("git", "clone", url, dir)
		if err := git.Run(); err != nil {
			errChannel <- err
			return
		}
	} else {
		git := exec.Command("git", "fetch")
		if err := git.Run(); err != nil {
			errChannel <- err
			return
		}
		update = true
	}

	// At the end, add dir path to buildChannel
	defer func() {
		buildChannel <- &PkgBuild{dir, conf.CacheDir, name, version, depends, makeDepends, optDepends, update, false}
	}()

	errChannel <- nil
}

// Passes arg to pacman -S
func pacmanSync(args []string, silent bool, deps bool) []error {
	if len(args) == 0 {
		return nil
	}

	errOut := []error{}
	for _, arg := range args {
		if !silent {
			output.Printf("Installing \033[1m\033[32m%s\033[39m\033[0m with \033[1mpacman\033[0m", arg)
		}
	}

	if deps {
		args = append([]string{"--asdeps"}, args...)
	}
	args = append([]string{"-S", "--noconfirm"}, args...)
	args = append([]string{"pacman"}, args...)

	cmd := exec.Command("sudo", args...)
	output.SetStd(cmd)
	if err := cmd.Run(); err != nil {
		errOut = append(errOut, err)
	}
	return errOut
}

type depBuild struct {
	name    string
	version string
	greater bool
}

var depCheckMap = make(map[string]bool)

func checkRecursively(pkg *PkgBuild, out, outMake, outOpts []PkgBuild) {
	newDeps, newMakeDeps, newOptDeps, _ := pkg.depCheck()
	for _, dep := range newDeps {
		depCheckMap[dep.name] = true
	}
	for _, dep := range newMakeDeps {
		depCheckMap[dep.name] = true
	}
	for _, dep := range newOptDeps {
		depCheckMap[dep.name] = true
	}
	out = append(out, newDeps...)
	outMake = append(outMake, newMakeDeps...)
	outOpts = append(outOpts, newOptDeps...)
}

// depCheck for AUR dependencies
// Downloads PKGBUILD's recursively
func (pkg *PkgBuild) depCheck() ([]PkgBuild, []PkgBuild, []PkgBuild, error) {
	// Dependencies
	deps := []depBuild{}
	for _, dep := range pkg.depends {
		// TODO: fix for specifiers
		deps = append(deps, parseDep(dep))
	}
	// Make Dependencies
	makeDeps := []depBuild{}
	for _, dep := range pkg.makeDepends {
		makeDeps = append(makeDeps, parseDep(dep))
	}

	optDeps := []depBuild{}
	for _, dep := range pkg.optDepends {
		optDeps = append(optDeps, parseDep(dep))
	}

	// Sync deps
	depNames := []string{}
	for _, dep := range deps {
		// Check if installed
		check := exec.Command("pacman", "-Qi", dep.name)
		if err := check.Run(); err != nil {
			// Probs not installed
			depNames = append(depNames, dep.name)
		}
	}

	// Sync makeDeps
	makeDepNames := []string{}
	for _, dep := range makeDeps {
		// Check if installed
		check := exec.Command("pacman", "-Qi", dep.name)
		if err := check.Run(); err != nil {
			// Probs not installed
			makeDepNames = append(makeDepNames, dep.name)
		}
	}

	// Sync optDeps
	optDepNames := []string{}
	for _, dep := range optDeps {
		// Check if installed
		check := exec.Command("pacman", "-Qi", dep.name)
		if err := check.Run(); err != nil {
			// Probs not installed
			optDepNames = append(optDepNames, dep.name)
		}
	}

	// Download func
	dload := func(errChannel chan error, buildChannel chan *PkgBuild, dep string) {
		repo, err := aur.Info([]string{dep})
		if err != nil {
			output.PrintErr("Dependencies error: %s", err)
		}
		if len(repo) > 0 {
			go aurDload("https://aur.archlinux.org/"+repo[0].PackageBase+".git", errChannel, buildChannel, repo[0].PackageBase, repo[0].Version, repo[0].Depends, repo[0].MakeDepends, repo[0].OptDepends)
		} else {
			// Not on the aur
			errChannel <- nil
			buildChannel <- &PkgBuild{name: dep, pacman: true}
		}
	}

	// Now, get PKGBUILDs
	// For deps
	errChannel := make(chan error, len(depNames))
	buildChannel := make(chan *PkgBuild, len(depNames))
	for _, dep := range depNames {
		dload(errChannel, buildChannel, dep)
	}

	// For makeDeps
	errChannelM := make(chan error, len(makeDepNames))
	buildChannelM := make(chan *PkgBuild, len(makeDepNames))
	for _, dep := range makeDepNames {
		dload(errChannelM, buildChannelM, dep)
	}

	// For optDeps
	errChannelO := make(chan error, len(optDepNames))
	buildChannelO := make(chan *PkgBuild, len(optDepNames))
	for _, dep := range optDepNames {
		dload(errChannelO, buildChannelO, dep)
	}

	out := []PkgBuild{}
	outMake := []PkgBuild{}
	outOpts := []PkgBuild{}

	// Collect deps
	for _i := 0; _i < len(depNames)*2; _i++ {
		select {
		case pkg := <-buildChannel:
			out = append(out, *pkg)
			// Map dependency tree
			if !pkg.pacman && !depCheckMap[pkg.name] {
				checkRecursively(pkg, out, outMake, outOpts)
			}
		case err := <-errChannel:
			if err != nil {
				output.PrintErr("Dependencies error: %s", err)
			}
		}
	}

	// Collect makeDeps
	for _i := 0; _i < len(makeDepNames)*2; _i++ {
		select {
		case pkg := <-buildChannelM:
			outMake = append(outMake, *pkg)
			// Map dependency tree
			if !pkg.pacman && !depCheckMap[pkg.name] {
				checkRecursively(pkg, out, outMake, outOpts)
			}
		case err := <-errChannelM:
			if err != nil {
				output.PrintErr("Dependencies error: %s", err)
			}
		}
	}

	// Collect optDeps
	for _i := 0; _i < len(optDepNames)*2; _i++ {
		select {
		case pkg := <-buildChannelO:
			outOpts = append(outOpts, *pkg)
			// Map dependency tree
			if !pkg.pacman {
				newDeps, newMakeDeps, newOptDeps, _ := pkg.depCheck()
				fmt.Println(newDeps, newMakeDeps, newOptDeps)
				out = append(out, newDeps...)
				outMake = append(outMake, newMakeDeps...)
				outOpts = append(outOpts, newOptDeps...)
			}
		case err := <-errChannelO:
			if err != nil {
				output.PrintErr("Dependencies error: %s", err)
			}
		}
	}

	// Filter deps for repition
	seen := map[string]bool{}
	seenMake := map[string]bool{}
	seenOpts := map[string]bool{}
	outF := []PkgBuild{}
	outMakeF := []PkgBuild{}
	outOptsF := []PkgBuild{}
	for _, pack := range out {
		if !seen[pack.name] {
			seen[pack.name] = true
			outF = append(outF, pack)
		}
	}

	for _, pack := range outMake {
		if !seenMake[pack.name] {
			seenMake[pack.name] = true
			outMakeF = append(outMakeF, pack)
		}
	}

	for _, pack := range outOpts {
		if !seenOpts[pack.name] {
			seenOpts[pack.name] = true
			outOptsF = append(outOptsF, pack)
		}
	}
	return outF, outMakeF, outOptsF, nil
}

// Get dependency syntax
func parseDep(dep string) depBuild {
	dep = strings.TrimSpace(dep)

	if strings.Contains(dep, "=<") {
		dep = strings.Split(dep, "=<")[0]
	} else if strings.Contains(dep, "=>") {
		dep = strings.Split(dep, "=>")[0]
	} else if strings.Contains(dep, ">") {
		dep = strings.Split(dep, ">")[0]
	} else if strings.Contains(dep, "<") {
		dep = strings.Split(dep, "<")[0]
	} else if strings.Contains(dep, "==") {
		dep = strings.Split(dep, "==")[0]
	} else if strings.Contains(dep, "=") {
		dep = strings.Split(dep, "=")[0]
	}
	return depBuild{name: dep}
}
