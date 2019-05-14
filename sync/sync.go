package sync

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	total    int64
	length   int64
	progress float64
}

// Read will override io.Reader's Read method
//
// Source: https://stackoverflow.com/questions/25645363/track-and-show-downloading-file-summary-in-percentage-go-lang#25645804
func (dl *Download) Read(p []byte) (int, error) {
	num, err := dl.Reader.Read(p)
	if num > 0 {
		dl.total += int64(num)
		percentage := float64(dl.total) / float64(dl.length) * float64(100)
		perInt := int(percentage / float64(10))
		out := fmt.Sprintf("%v", perInt)
		if percentage-dl.progress > 2 {
			fmt.Fprintf(os.Stderr, out)
		}
	}
	return num, err
}

// Sync from the AUR first, then other configured repos.
//
// This checks each package param individually
func Sync(packages []string) error {
	// TODO: Check with config
	errChannel := make(chan error, len(packages))
	for _, p := range packages {
		// Multithreaded downloads
		go func(p string) {
			repo, err := aur.Info([]string{p})
			if err != nil {
				errChannel <- err
			}
			if len(repo) > 0 {
				aurDload("https://aur.archlinux.org"+repo[0].URLPath, repo[0].Name+repo[0].Version+".tar.gz", errChannel)
			}
		}(p)
	}

	for _i := 0; _i < len(packages); _i++ {
		err := <-errChannel
		if err != nil {
			return err
		}
	}

	return nil
}

// Download an AUR package to cache
func aurDload(url string, fileName string, errChannel chan error) {
	resp, err := http.Get(url)
	if err != nil {
		errChannel <- err
		return
	}
	defer resp.Body.Close()

	conf := config.GetConfig()
	file := filepath.Join(conf.CacheDir, fileName)

	out, err := os.Create(file)
	if err != nil {
		errChannel <- err
		return
	}
	defer out.Close()

	_, errC := io.Copy(out, resp.Body)
	if errC != nil {
		errChannel <- errC
		return
	}

	errChannel <- nil
}
