package sync

import (
	"fmt"

	"github.com/ericm/yup/config"
	"github.com/mikkeloscar/aur"
)

// func Search(terms ...string) error {
// 	fmt.Println(aur.AURURL)
// 	return nil
// }

// Sync from the AUR first, then other configured repos.
//
// This checks each package param individually
func Sync(packages []string) error {
	// TODO: Check with config
	errChannel := make(chan error, len(packages))
	for _, p := range packages {
		repo, err := aur.Info([]string{p})
		if err != nil {
			errChannel <- err
		}
		if len(repo) > 0 {
			// Multithreaded downloads
			go aurDload("https://aur.archlinux.org"+repo[0].URLPath, errChannel)
		}
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
func aurDload(url string, err chan error) {
	conf := config.GetConfig()
	fmt.Println(conf.CacheDir)
	err <- nil
}
