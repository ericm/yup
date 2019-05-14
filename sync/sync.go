package sync

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

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
	file string
	dir  string
}

// Read will override io.Reader's Read method
//
// Source: https://stackoverflow.com/questions/25645363/track-and-show-downloading-file-summary-in-percentage-go-lang#25645804
func (dl *Download) Read(p []byte) (int, error) {
	num, err := dl.Reader.Read(p)
	if num > 0 {
		dl.total += int64(num)
		st := ""
		// Removes previous status message
		if dl.count > 0 {
			st = "\033[F\033[K"
		}
		fmt.Printf("%sDownloaded: %vB\n", st, dl.total)
		dl.count++
	}
	return num, err
}

// Sync from the AUR first, then other configured repos.
//
// This checks each package param individually
func Sync(packages []string) error {
	// Create channels for goroutines
	errChannel := make(chan error, len(packages))
	buildChannel := make(chan *pkgBuild, len(packages))
	for _, p := range packages {
		// Multithreaded downloads
		go func(p string) {
			repo, err := aur.Info([]string{p})
			if err != nil {
				errChannel <- err
			}
			if len(repo) > 0 {
				aurDload("https://aur.archlinux.org"+repo[0].URLPath, repo[0].Name+repo[0].Version+".tar.gz", errChannel, buildChannel)
			}
		}(p)
	}

	for _i := 0; _i < len(packages)*2; _i++ {
		// Check for both error and build Channels
		select {
		case err := <-errChannel:
			if err != nil {
				return err
			}
		case pkg := <-buildChannel:
			os.Chdir(pkg.dir)
			cmd := exec.Command("tar", "-zxvf", pkg.file)
			if err := cmd.Run(); err != nil {
				fmt.Fprint(os.Stderr, err)
			}
		}

	}

	return nil
}

// Download an AUR package to cache
func aurDload(url string, fileName string, errChannel chan error, buildChannel chan *pkgBuild) {
	conf := config.GetConfig()
	file := filepath.Join(conf.CacheDir, fileName)
	// At the end, add file path to buildChannel
	defer func() {
		buildChannel <- &pkgBuild{file, conf.CacheDir}
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
