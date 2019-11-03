package clean

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ericm/yup/config"
	"github.com/ericm/yup/output"
)

// Clean unused packages and delete cache
func Clean() error {
	if err := Aur(); err != nil {
		return err
	}

	output.Printf("Clearing pacman cache")
	pc := exec.Command("sudo", "pacman", "-Sc")
	output.SetStd(pc)
	if err := pc.Run(); err != nil {
		return err
	}

	// Clear unused packs
	output.Printf("Finding unused dependencies")
	pac := exec.Command("pacman", "-Qtdq")
	var packs []string
	if out, err := pac.Output(); err == nil {
		for _, pack := range strings.Split(string(out), "\n") {
			if len(pack) > 1 {
				packs = append(packs, pack)
			}
		}
	}

	// Remove
	packs = append([]string{"pacman", "-Rns"}, packs...)
	rem := exec.Command("sudo", packs...)
	output.SetStd(rem)
	if err := rem.Run(); err != nil {
		return err
	}

	return nil
}

// Aur clears AUR cache
func Aur() error {
	cache := config.GetConfig().CacheDir
	dirs, err := ioutil.ReadDir(cache)
	if err != nil {
		return err
	}

	output.Printf("Clearing your AUR cache...")
	os.Chdir(cache)
	for _, dir := range dirs {
		if dir.IsDir() {
			os.Chdir(dir.Name())
			// Delete big bad files. git clean -f will only rm source files
			gitrm := exec.Command("git", "clean", "-f")
			if err := gitrm.Run(); err != nil {
				return err
			}
		}
		os.Chdir(cache)
	}
	return nil
}
