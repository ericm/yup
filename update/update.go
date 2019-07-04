package update

import (
	"fmt"
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
	cmd := exec.Command("pacman", "-Qm")
	inp, err := cmd.Output()
	if err != nil {
		return err
	}

	var updates []installedPack

	packStr := strings.Split(string(inp), "\n")
	for _, pack := range packStr {
		p := strings.Split(pack, " ")
		pack := installedPack{p[0], p[1]}
		aurPack, errAur := aur.Info([]string{pack.name})
		if errAur != nil {
			output.PrintErr("%s", errAur)
		}
		if len(aurPack) > 0 && aurPack[0].Version != pack.version {
			fmt.Println(pack.name)
		}
	}
	fmt.Print(updates)

	return nil
}
