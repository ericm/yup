package clean

import (
	"github.com/ericm/yup/config"
	"github.com/ericm/yup/output"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Clean unused packages and delete cache
func Clean() error {
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
			// Delete big bad files
			tarGz, _ := filepath.Glob("*.tar.gz")
			for _, tar := range tarGz {
				if err := os.Remove(tar); err != nil {
					output.PrintErr("%s", err)
				}
			}
			tarXz, _ := filepath.Glob("*.pkg.tar.xz")
			for _, tar := range tarXz {
				if err := os.Remove(tar); err != nil {
					output.PrintErr("%s", err)
				}
			}

			if err := os.RemoveAll("pkg"); err != nil {
				output.PrintErr("%s", err)
			}
			if err := os.RemoveAll("src"); err != nil {
				output.PrintErr("%s", err)
			}
		}
		os.Chdir(cache)
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
