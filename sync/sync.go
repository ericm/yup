package sync

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ericm/yup/output"
	"github.com/mikkeloscar/aur"

	"fmt"

	"github.com/ericm/yup/config"
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

type pkgBuild struct {
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
	buildChannel := make(chan *pkgBuild, len(packages))

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
					aurDload("https://aur.archlinux.org/"+repo[0].Name+".git", errChannel, buildChannel, repo[0].Name, repo[0].Version, repo[0].Depends, repo[0].MakeDepends)
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
		sync := pacmanSync(pacmanArgs, false)
		for _, s := range sync {
			if s != nil {
				return s
			}
		}
	}

	return nil
}

// Install the pkgBuild
// assuming repo is now cloned or fetched
func (pkg *pkgBuild) Install(silent bool) error {
	if !silent {
		output.Printf("Installing \033[1m\033[32m%s\033[39m\033[2m v%s\033[0m from the AUR", pkg.name, pkg.version)
	}

	// Install from the AUR
	os.Chdir(filepath.Join(pkg.dir, pkg.name))

	scanner := bufio.NewReader(os.Stdin)
	if !silent {
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
				var diff *exec.Cmd
				if pkg.update {
					diff = exec.Command("git", "diff", "master", "origin/master")
				} else {
					diff = exec.Command("git", "diff", "@~..@")
				}

				cmds = append(cmds, diff)

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

	// Then merge if update
	if pkg.update {
		merge := exec.Command("git", "merge", "origin/master")
		merge.Run()
	}

	// Make / Install the package
	pkg.dir = filepath.Join(pkg.dir, pkg.name)

	remMakes := false
	if !silent {
		// Check for dependencies
		deps, makeDeps, err := pkg.depCheck()
		if err != nil {
			return err
		}

		if len(deps) > 0 {
			output.Printf("Found uninstalled Dependencies:")
			fmt.Print("    ")
			for _, dep := range deps {
				fmt.Printf("%s  ", dep.name)
			}
			fmt.Print("\n")
		}

		if len(makeDeps) > 0 {
			output.Printf("Found uninstalled Make Dependencies:")
			fmt.Print("    ")
			for _, dep := range makeDeps {
				fmt.Printf("%s  ", dep.name)
			}
			fmt.Print("\n")
			output.PrintIn("Remove Make Dependencies after install? (y/N)")
			rem, _ := scanner.ReadString('\n')
			switch strings.TrimSpace(strings.ToLower(rem[:1])) {
			case "y":
				remMakes = true
				break
			}
		}

		if len(deps) > 0 {
			output.Printf("Installing dependencies")
			// Install deps packages
			for _, dep := range deps {
				if dep.pacman {
					// Install from pacman in silent mode
					pacmanSync([]string{dep.name}, true)
				} else {
					// Install using Install in silent mode
					err := dep.Install(true)
					if err != nil {
						output.PrintErr("Dep Install error:")
						return err
					}
				}
			}
		}

		if len(makeDeps) > 0 {
			output.Printf("Installing dependencies")
			// Install makeDeps packages
			for _, dep := range makeDeps {
				if dep.pacman {
					// Install from pacman in silent mode
					pacmanSync([]string{dep.name}, true)
				} else {
					// Install using Install in silent mode
					err := dep.Install(true)
					if err != nil {
						output.PrintErr("Dep Install error:")
						return err
					}
				}
			}

			// At end, remove make packs as necessary
			if remMakes {
				defer func(depM []pkgBuild) {
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

	}

	// Now, Install the actual package
	os.Chdir(pkg.dir)
	cmdMake := exec.Command("makepkg", "-si")
	// Pipe to stdout, etc
	output.SetStd(cmdMake)
	if err := cmdMake.Run(); err != nil {
		return err
	}

	return nil
}

// Download an AUR package to cache
func aurDload(url string, errChannel chan error, buildChannel chan *pkgBuild, name string, version string, depends []string, makeDepends []string) {
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
		buildChannel <- &pkgBuild{dir, conf.CacheDir, name, version, depends, makeDepends, update, false}
	}()

	errChannel <- nil
}

// Passes arg to pacman -S
func pacmanSync(args []string, silent bool) []error {
	errOut := []error{}
	for _, arg := range args {
		if !silent {
			output.Printf("Installing \033[1m\033[32m%s\033[39m\033[0m with \033[1mpacman\033[0m", arg)
		}
		cmd := exec.Command("sudo", "pacman", "-S", arg)
		output.SetStd(cmd)
		if err := cmd.Run(); err != nil {
			errOut = append(errOut, err)
		}
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
func (pkg *pkgBuild) depCheck() ([]pkgBuild, []pkgBuild, error) {
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
	dload := func(errChannel chan error, buildChannel chan *pkgBuild, dep string) {
		repo, err := aur.Info([]string{dep})
		if err != nil {
			output.PrintErr("Dependencies error: %s", err)
		}
		if len(repo) > 0 {
			go aurDload("https://aur.archlinux.org/"+repo[0].Name+".git", errChannel, buildChannel, repo[0].Name, repo[0].Version, repo[0].Depends, repo[0].MakeDepends)
		} else {
			// Not on the aur
			errChannel <- nil
			buildChannel <- &pkgBuild{name: dep, pacman: true}
		}
	}

	// Now, get PKGBUILDs
	// For deps
	errChannel := make(chan error, len(depNames))
	buildChannel := make(chan *pkgBuild, len(depNames))
	for _, dep := range depNames {
		dload(errChannel, buildChannel, dep)
	}

	// For makeDeps
	errChannelM := make(chan error, len(makeDepNames))
	buildChannelM := make(chan *pkgBuild, len(makeDepNames))
	for _, dep := range makeDepNames {
		dload(errChannelM, buildChannelM, dep)
	}

	out := []pkgBuild{}
	outMake := []pkgBuild{}
	// Collect deps
	for _i := 0; _i < len(depNames)*2; _i++ {
		select {
		case pkg := <-buildChannel:
			out = append(out, *pkg)
			// Map dependency tree
			if !pkg.pacman {
				newDeps, newMakeDeps, _ := pkg.depCheck()
				out = append(out, newDeps...)
				outMake = append(outMake, newMakeDeps...)
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
		case pkg := <-buildChannel:
			outMake = append(outMake, *pkg)
			// Map dependency tree
			if !pkg.pacman {
				newDeps, newMakeDeps, _ := pkg.depCheck()
				out = append(out, newDeps...)
				outMake = append(outMake, newMakeDeps...)
			}
		case err := <-errChannel:
			if err != nil {
				output.PrintErr("Dependencies error: %s", err)
			}
		}
	}

	return out, outMake, nil
}

// Get dependency syntax
func parseDep(dep string) depBuild {
	dep = strings.TrimSpace(dep)

	if strings.Contains(dep, "=<") {

	}

	return depBuild{name: dep}
}
