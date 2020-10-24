package update

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ericm/yup/config"
	"github.com/ericm/yup/output"
	"github.com/ericm/yup/sync"
	"github.com/mikkeloscar/aur"
)

// Installed Packages representation
type installedPack struct {
	name,
	version,
	newVersion string
}

// Update runs system update from repos
func Update() error {
	flg := "-Syu"
	if strings.Contains(os.Args[len(os.Args)-1], "yy") {
		flg = "-Syyu"
	}
	output.Printf("Updating from local repositories")
	cmd := exec.Command("sudo", "pacman", flg)
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
	var outdated []installedPack

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
		if len(aurPack) > 0 {
			if newerVersion(pack.version, aurPack[0].Version) {
				pack.newVersion = aurPack[0].Version
				updates = append(updates, pack)
			} else if pack.version != aurPack[0].Version {
				// Package must be newer than AUR
				pack.newVersion = aurPack[0].Version
				outdated = append(outdated, pack)
			}
		}
	}

	scanner := bufio.NewReader(os.Stdin)
	if len(outdated) > 0 {
		fmt.Print("\n")
		output.Printf("Found %d local package(s) that are newer than their AUR package", len(outdated))
		for _, pack := range outdated {
			fmt.Printf("    \033[1m%s\033[0m  \033[95m%s\033[0m has AUR version \033[95m%s\033[0m\n", pack.name, pack.version, pack.newVersion)
		}
	}
	fmt.Print("\n")
	if len(updates) == 0 {
		output.Printf("Found no AUR packages to update")
		return nil
	}
	output.Printf("Found %d AUR package(s) to update:", len(updates))
	for i, pack := range updates {
		fmt.Printf("    %-3d \033[1m%s\033[0m \033[91m%s\033[0m -> \033[92m%s\033[0m\n", i+1, pack.name, pack.version, pack.newVersion)
	}

	output.PrintIn("Packages not to install? (eg: 1 2 3, 1-3 or ^4)")

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
	}
	return sync.Sync(syncUp, true, false)
}

func newerVersion(oldVersion, newVersion string) bool {
	oldVer := strings.Split(oldVersion, "-")
	newVer := strings.Split(newVersion, "-")
	if len(oldVer) > 1 && len(newVer) > 1 {
		if len(oldVer[0]) > 7 {
			return oldVersion != newVersion // Likely commit hashed
		}
		// For rXX
		rSplitO := strings.SplitAfter(oldVer[0], "r")
		rSplitN := strings.SplitAfter(newVer[0], "r")
		if len(rSplitO) > 0 && len(rSplitN) > 0 {
			oldVer[0] = rSplitO[0]
			newVer[0] = rSplitN[0]
		}
		// Get rel
		if oldVer[0] == newVer[0] {
			relOld, _ := strconv.Atoi(oldVer[1])
			relNew, _ := strconv.Atoi(newVer[1])
			if len(rSplitO) > 1 && len(rSplitN) > 1 {
				rSplitOld, _ := strconv.Atoi(rSplitO[1])
				rSplitNew, _ := strconv.Atoi(rSplitN[1])
				if rSplitOld < rSplitNew {
					return true
				}
			}
			if relOld < relNew {
				return true
			}
			return false
		}
		// Get version diff
		dotOld := strings.Split(oldVer[0], ".")
		dotNew := strings.Split(newVer[0], ".")
		var m int
		if len(dotOld) > len(dotNew) {
			m = len(dotNew)
		} else {
			m = len(dotOld)
		}
		for i := 0; i < m; i++ {
			// Get ints
			oldNum, errOld := strconv.Atoi(dotOld[i])
			newNum, errNew := strconv.Atoi(dotNew[i])
			if errOld == nil && errNew == nil {
				if newNum > oldNum {
					return true
				}
			} else {
				if dotNew[i] > dotOld[i] {
					return true
				}
			}
		}
	}

	return false
}
