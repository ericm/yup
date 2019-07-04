package update

import (
	"github.com/ericm/yup/output"
	"github.com/mikkeloscar/aur"
	"os/exec"
	"strings"
)

// Installed Packages representation
type installedPack struct {
	name,
	version string
}

// Update runs system update from repos
func Update() error {
	output.Printf("Updating from local repositories")
	cmd := exec.Command("sudo", "pacman", "-Syyu")
	output.SetStd(cmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Launch AUR update
	return AurUpdate()
}

// AurUpdate checks for update in the AUR
func AurUpdate() error {
	// Filter installed packages

	// Get output of pacman -Q
	cmd := exec.Command("pacman", "-Q")
	inp, err := cmd.Output()
	if err != nil {
		return err
	}

	packStr := strings.Split(string(inp), "\n")
	for _, pack := range packStr {
		p := strings.Split(pack, " ")
		pack := installedPack{p[0], p[1]}
		pkg, _ := aur.Info([]string{pack.name})
		if len(pkg) > 0 {
			// Aur pack found
			if pkg[0].Version != pack.version {
				// Version changed
				// Update
			}
		}
	}

	return nil
}
