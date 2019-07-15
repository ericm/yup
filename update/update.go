package update

import (
	"bufio"
	"fmt"
	"github.com/ericm/yup/config"
	"github.com/ericm/yup/output"
	"github.com/ericm/yup/sync"
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
	output.Printf("Checking for AUR updates...")

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

	if len(updates) == 0 {
		output.Printf("Found no AUR packages to update")
		return nil
	}
	output.Printf("Found %d AUR package(s) to update:", len(updates))
	for i, pack := range updates {
		fmt.Printf("    %-3d \033[1m%s\033[0m \033[91m%s\033[0m -> \033[92m%s\033[0m\n", i+1, pack.name, pack.version, pack.newVersion)
	}

	output.PrintIn("Packages not to install? (eg: 1 2 3, 1-3 or ^4)")

	scanner := bufio.NewReader(os.Stdin)
	not, _ := scanner.ReadString('\n')

	syncUp := []string{}

	seen := map[int]bool{}
	for _, s := range strings.Split(strings.TrimSpace(not), " ") {
		// 1-3
		if strings.Contains(s, "-") {
			if spl := strings.Split(s, "-"); len(spl) == 2 {
				// Get int vals for range
				firstT, errF := strconv.Atoi(spl[0])
				secondT, errS := strconv.Atoi(spl[1])
				if errF == nil && errS == nil {
					// Convert range from visual representation
					first := len(updates) - firstT
					second := len(updates) - secondT
					// Filter
					for i := second; i <= first; i++ {
						seen[i] = true
					}
				}
			}
			continue
		}
		// ^4
		if strings.Contains(s, "^") {
			if num, err := strconv.Atoi(s[1:]); err == nil {
				// Filter for the number
				for i := range updates {
					if i == num {
						seen[i] = true
						continue
					}
				}
			}
			continue
		}

		if num, err := strconv.Atoi(s); err == nil {
			if !seen[num] {
				seen[num] = true
			} else {
			}
		}
	}

	for i, u := range updates {
		if !seen[i+1] {
			syncUp = append(syncUp, u.name)
		}
	}

	if config.GetConfig().UserFile.SilentUpdate {
		return sync.Sync(syncUp, true, true)
	} else {
		return sync.Sync(syncUp, true, false)
	}
}
