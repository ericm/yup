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
	version,
	newVersion string
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
		if len(p) < 2 {
			continue
		}
		pack := installedPack{name: p[0], version: p[1]}
		aurPack, errAur := aur.Info([]string{pack.name})
		if errAur != nil {
			output.PrintErr("%s", errAur)
		}
		if len(aurPack) > 0 && aurPack[0].Version != pack.version {
			pack.newVersion = aurPack[0].Version
			updates = append(updates, pack)
		}
	}

	output.Printf("Found %d AUR package(s) to update:", len(updates))
	for i, pack := range updates {
		fmt.Printf("    \033[1m%d %s\033[0m \033[91m%s\033[0m -> \033[92m%s\033[0m\n", i+1, pack.name, pack.version, pack.newVersion)
	}

	return nil
}
