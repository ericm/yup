package sync

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

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

// Read will override io.Reader's Read method
//
// Source: https://stackoverflow.com/questions/25645363/track-and-show-downloading-file-summary-in-percentage-go-lang#25645804
func (dl *Download) Read(p []byte) (int, error) {
	num, err := dl.Reader.Read(p)
	if num > 0 {
		dl.total += int64(num)
		// st := ""
		// // Removes previous status message
		// if dl.count > 0 {
		// 	st = "\033[F\033[K"
		// }
		// fmt.Printf("%sDownloaded: %vB\n", st, dl.total)
		dl.count++
	}
	return num, err
}

// Sync from the AUR first, then other configured repos.
//
// This checks each package param individually
func Sync(packages []string) error {
	// Create channels for goroutines
	// Step 1: Check AUR
	output.Printf("Checking the \033[1mAUR\033[0m")
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
					aurDload("https://aur.archlinux.org"+repo[0].URLPath, repo[0].Name+repo[0].Version+".tar.gz", errChannel, buildChannel, repo[0].Name, repo[0].Version)
				} else {
					errChannel <- output.Errorf("Didn't find an \033[1mAUR\033[0m package for \033[1m\033[32m%s\033[39m\033[0m, searching other repos", p)
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

				// Untar the package
				os.Chdir(pkg.dir)
				cmdTar := exec.Command("tar", "-zxvf", pkg.file)
				if err := cmdTar.Run(); err != nil {
					return err
				}

				// TODO: View PKGBUILD

				// Make / Install the package
				pkg.dir = filepath.Join(pkg.dir, pkg.name)
				os.Chdir(pkg.dir)
				cmdMake := exec.Command("makepkg", "-si")
				// Pipe to stdout, etc
				cmdMake.Stdout, cmdMake.Stdin, cmdMake.Stderr = os.Stdout, os.Stdin, os.Stderr
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
func aurDload(url string, fileName string, errChannel chan error, buildChannel chan *pkgBuild, name string, version string) {
	// TODO: Check in cache
	conf := config.GetConfig()
	file := filepath.Join(conf.CacheDir, fileName)
	// At the end, add file path to buildChannel
	defer func() {
		buildChannel <- &pkgBuild{file, conf.CacheDir, name, version}
	}()

	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		errChannel <- err
		return
	}
	defer resp.Body.Close()

	download := &Download{Reader: resp.Body, count: 0}
	body, err := ioutil.ReadAll(download)
	if err != nil {
		errChannel <- err
		return
	}

	out, err := os.Create(file)
	if err != nil {
		errChannel <- err
		return
	}
	defer out.Close()

	_, errC := out.Write(body)
	if errC != nil {
		errChannel <- errC
		return
	}

	errChannel <- nil
}

// Passes arg to pacman -S
func pacmanSync(args []string) []error {
	errOut := []error{}
	for _, arg := range args {
		output.Printf("Installing \033[1m\033[32m%s\033[39m\033[0m with \033[1mpacman\033[0m", arg)
		cmd := exec.Command("sudo", "pacman", "-S", arg)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			errOut = append(errOut, err)
		}
	}

	return errOut
}
