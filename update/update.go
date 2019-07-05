package update

import (
	"bufio"
	"fmt"
	"github.com/ericm/yup/output"
	"github.com/mikkeloscar/aur"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Installed Packages representation
type installedPack struct {
	name,
	version,
	newVersion string
	noInstall bool
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
		fmt.Printf("    %d \033[1m%s\033[0m \033[91m%s\033[0m -> \033[92m%s\033[0m\n", i+1, pack.name, pack.version, pack.newVersion)
	}

	output.PrintIn("Packages not to install? (eg: 1 2 3, 1-3 or ^4)")

	scanner := bufio.NewReader(os.Stdin)
	not, _ := scanner.ReadString('\n')

	seen := map[int]bool{}
	for _, s := range strings.Split(strings.TrimSpace(not), " ") {
		// 1-3
		if strings.Contains(s, "-") {
			continue
		}
		// ^4
		if strings.Contains(s, "^") {
			continue
		}

		if num, err := strconv.Atoi(s); err == nil && !seen[num] {
			seen[num] = true
			updates[num].noInstall = true
		}
	}

	return nil
}
