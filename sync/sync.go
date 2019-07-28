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

	"github.com/ericm/yup/config"
	"strconv"
)

// func Search(terms ...string) error {
// 	fmt.Println(aur.AURURL)
// 	return nil
// }

// Download wrapper for io.Reader
type Download struct {
	io.Reader
	total int64
	count int
}

// Represents a PkgBuild
type PkgBuild struct {
	file        string
	dir         string
	name        string
	version     string
	depends     []string
	makeDepends []string
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
					aurDload("https://aur.archlinux.org/"+repo[0].PackageBase+".git", errChannel, buildChannel, repo[0].PackageBase, repo[0].Version, repo[0].Depends, repo[0].MakeDepends)
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
				if err := pkg.Install(silent); err != nil {
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

// Install the pkgBuild
// assuming repo is now cloned or fetched
func (pkg *PkgBuild) Install(silent bool) error {
	output.Printf("Installing \033[1m\033[32m%s\033[39m\033[2m %s\033[0m from the AUR", pkg.name, pkg.version)

	// Install from the AUR
	os.Chdir(filepath.Join(pkg.dir, pkg.name))

	scanner := bufio.NewReader(os.Stdin)
	if pkg.update {
		merge := exec.Command("git", "merge", "origin/master")
		merge.Run()
	}
	if !silent {
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

	remMakes := false
	if !silent {
		// Check for dependencies
		output.Printf("Checking for dependencies")
		deps, makeDeps, err := pkg.depCheck()
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
			output.PrintIn("Numbers of packages not to install? (eg: 1 2 3, 1-3 or ^4)")
			depRem, _ := scanner.ReadString('\n')

			// Parse input
			ParseNumbers(depRem, &deps)
		}

		if len(makeDeps) > 0 {
			output.Printf("Found uninstalled Make Dependencies:")
			fmt.Print("    ")
			for i, dep := range makeDeps {
				fmt.Printf("\033[1m%d\033[0m %s  ", i+1, dep.name)
			}
			fmt.Print("\n")

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

		// Gather packages
		aurInstall := []PkgBuild{}
		pacInstall := []string{}

		if len(deps) > 0 {
			// Install deps packages
			for _, dep := range deps {
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
			for _, dep := range makeDeps {
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

		if len(pacInstall) > 0 || len(aurInstall) > 0 {
			output.Printf("Installing Dependencies")
		}
		// Pacman deps
		if err := pacmanSync(pacInstall, true, true); err != nil {
			//output.PrintErr("%s", err)
		}
		// Aur deps
		for _, dep := range aurInstall {
			err := dep.Install(true)
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

	os.Chdir(pkg.dir)

	// Get PGP Keys
	info, err := srcinfo.ParseFile(".SRCINFO")
	if err != nil {
		return err
	}
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
func aurDload(url string, errChannel chan error, buildChannel chan *PkgBuild, name string, version string, depends []string, makeDepends []string) {
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
		buildChannel <- &PkgBuild{dir, conf.CacheDir, name, version, depends, makeDepends, update, false}
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

// depCheck for AUR dependencies
// Downloads PKGBUILD's recursively
func (pkg *PkgBuild) depCheck() ([]PkgBuild, []PkgBuild, error) {
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

	// Download func
	dload := func(errChannel chan error, buildChannel chan *PkgBuild, dep string) {
		repo, err := aur.Info([]string{dep})
		if err != nil {
			output.PrintErr("Dependencies error: %s", err)
		}
		if len(repo) > 0 {
			go aurDload("https://aur.archlinux.org/"+repo[0].PackageBase+".git", errChannel, buildChannel, repo[0].PackageBase, repo[0].Version, repo[0].Depends, repo[0].MakeDepends)
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

	out := []PkgBuild{}
	outMake := []PkgBuild{}
	// Collect deps
	for _i := 0; _i < len(depNames)*2; _i++ {
		select {
		case pkg := <-buildChannel:
			out = append(out, *pkg)
			// Map dependency tree
			if !pkg.pacman {
				newDeps, newMakeDeps, _ := pkg.depCheck()
				out = append(newDeps, out...)
				outMake = append(newMakeDeps, outMake...)
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
			if !pkg.pacman {
				newDeps, newMakeDeps, _ := pkg.depCheck()
				out = append(out, newDeps...)
				outMake = append(outMake, newMakeDeps...)
			}
		case err := <-errChannelM:
			if err != nil {
				output.PrintErr("Dependencies error: %s", err)
			}
		}
	}

	// Filter deps for repition
	seen := map[string]bool{}
	seenMake := map[string]bool{}
	outF := []PkgBuild{}
	outMakeF := []PkgBuild{}
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

	return out, outMake, nil
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
