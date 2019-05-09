package sync

import (
	"fmt"

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
	for _, p := range packages {
		repo, err := aur.Info([]string{p})
		if err != nil {
			return err
		}
		if len(repo) > 0 {
			fmt.Println(repo[0].ID)
		}

	}

	fmt.Println("Sync", aur.AURURL)

	return nil
}
