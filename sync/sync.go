package sync

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ericm/yup/output"

	"fmt"

	"github.com/ericm/yup/config"
	"github.com/mikkeloscar/aur"
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
	file    string
	dir     string
	name    string
	version string
}

// Sync from the AUR first, then other configured repos.
//
// This checks each package param individually
func Sync(packages []string) error {
	// Create channels for goroutines
	// Step 1: Check AUR
	errChannel := make(chan error, len(packages))
	buildChannel := make(chan *pkgBuild, len(packages))

	pacmanArgs := []string{}

	for _, p := range packages {
		// Multithreaded downloads
		go func(p string) {
			repo, err := aur.Info([]string{p})
			if err != nil {
				errChannel <- err
			} else {
				if len(repo) > 0 {
					aurDload("https://aur.archlinux.org/"+repo[0].Name+".git", errChannel, buildChannel, repo[0].Name, repo[0].Version)
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
			if err != nil {
				fmt.Print(err)
			}
		case pkg := <-buildChannel:
			if pkg != nil {
				output.Printf("Installing \033[1m\033[32m%s\033[39m\033[2m v%s\033[0m from the AUR", pkg.name, pkg.version)

				// Install from the AUR
				os.Chdir(filepath.Join(pkg.dir, pkg.name))

			Pkgbuild:
				scanner := bufio.NewReader(os.Stdin)
				output.PrintIn("View the PKGBUILD? (y/N)")
				out, _ := scanner.ReadString('\n')

				switch strings.ToLower(out[:1]) {
				case "y":
					showPkg := exec.Command("cat", "PKGBUILD")
					output.SetStd(showPkg)
					if err := showPkg.Run(); err != nil {
						return err
					}
				Diffs:
					output.PrintIn("View Diffs? (y/N)")
					diffs, _ := scanner.ReadString('\n')
					switch strings.ToLower(diffs[:1]) {
					case "y":
						// Use git diff @~..@
						diff := exec.Command("git", "diff", "@~..@")
						output.SetStd(diff)
						if err := diff.Run(); err != nil {
							return err
						}
					Edit:
						// Finally, ask if they want to edit the PKGBUILD
						output.PrintIn("Edit PKGBUILD? (y/N)")
						edit, _ := scanner.ReadString('\n')
						switch strings.ToLower(edit[:1]) {
						case "y":
							// Check for EDITOR
							editor := os.Getenv("EDITOR")
							if len(editor) == 0 {
								// Ask for editor
								output.PrintIn("No EDITOR environment variable set. Enter editor")
								newEditor, _ := scanner.ReadString('\n')
								editor = newEditor[:len(newEditor)-1]
							}

							editPkg := exec.Command(editor, "PKGBUILD")
							output.SetStd(editPkg)
							if err := editPkg.Run(); err != nil {
								return err
							}
							break
						case "n":
						case "\n":
							break
						default:
							output.PrintErr("Please press N or Y")
							goto Edit
						}
						break
					case "n":
					case "\n":
						break
					default:
						output.PrintErr("Please press N or Y")
						goto Diffs
					}

					break
				case "n":
				case "\n":
					break
				default:
					output.PrintErr("Please press N or Y")
					goto Pkgbuild
				}

				// Make / Install the package
				pkg.dir = filepath.Join(pkg.dir, pkg.name)
				os.Chdir(pkg.dir)
				cmdMake := exec.Command("makepkg", "-si")
				// Pipe to stdout, etc
				output.SetStd(cmdMake)
				if err := cmdMake.Run(); err != nil {
					return err
				}
				output.PrintL()
			}

		}

	}

	// Now check pacman for unresolved args in pacmanArgs
	if len(pacmanArgs) > 0 {
		sync := pacmanSync(pacmanArgs)
		for _, s := range sync {
			if s != nil {
				return s
			}
		}
	}

	return nil
}

// Download an AUR package to cache
func aurDload(url string, errChannel chan error, buildChannel chan *pkgBuild, name string, version string) {
	// TODO: Check in cache
	conf := config.GetConfig()
	dir := filepath.Join(conf.CacheDir, name)
	// At the end, add dir path to buildChannel
	defer func() {
		buildChannel <- &pkgBuild{dir, conf.CacheDir, name, version}
	}()

	// Check if git repo is clones
	if os.IsNotExist(os.Chdir(dir)) {
		git := exec.Command("git", "clone", url, dir)
		if err := git.Run(); err != nil {
			errChannel <- err
			return
		}
	} else {
		git := exec.Command("git", "pull")
		if err := git.Run(); err != nil {
			errChannel <- err
			return
		}
	}

	errChannel <- nil
}

// Passes arg to pacman -S
func pacmanSync(args []string) []error {
	errOut := []error{}
	for _, arg := range args {
		output.Printf("Installing \033[1m\033[32m%s\033[39m\033[0m with \033[1mpacman\033[0m", arg)
		cmd := exec.Command("sudo", "pacman", "-S", arg)
		output.SetStd(cmd)
		if err := cmd.Run(); err != nil {
			errOut = append(errOut, err)
		}
	}

	return errOut
}
