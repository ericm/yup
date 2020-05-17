package sync

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/Jguer/go-alpm"
	"github.com/ericm/yup/output"
)

// Remove a package
func Remove(name string) error {
	handle, err := alpm.Initialize("/", "/var/lib/pacman")
	if err != nil {
		return err
	}
	db, err := handle.LocalDB()
	if err != nil {
		return err
	}
	deps := getRequiredBy(db, name, map[string]bool{name: true})
	if len(deps) > 0 {
		scanner := bufio.NewReader(os.Stdin)
		output.Printf("These packages require %s:", name)
		fmt.Print("    ")
		for i, dep := range deps {
			fmt.Printf("\033[1m%d\033[0m %s  ", i+1, dep)
		}
		fmt.Println()
		output.PrintIn("Numbers of packages NOT to remove? (eg: 1 2 3, 1-3 or ^4)")
		depRem, _ := scanner.ReadString('\n')

		// Parse input
		ParseNumbersStr(depRem, &deps)
		if len(deps) > 0 {
			output.Printf("Removing packages requiring %s:", name)
			for _, dep := range deps {
				cmd := exec.Command("sudo", "pacman", "-R", "--noconfirm", dep)
				output.SetStd(cmd)
				if err := cmd.Run(); err != nil {
					output.PrintErr(err.Error())
				}
			}
		}
	}
	output.Printf("Removing %s:", name)
	cmd := exec.Command("sudo", "pacman", "-R", name)
	output.SetStd(cmd)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getRequiredBy(db *alpm.DB, name string, seen map[string]bool) []string {
	pkg := db.Pkg(name)
	if pkg == nil {
		return []string{}
	}
	deps := pkg.ComputeRequiredBy()
	if len(deps) == 0 {
		return []string{}
	}
	currDeps := []string{}
	for _, dep := range deps {
		if !seen[dep] {
			currDeps = append(currDeps, dep)
		} else {
			seen[dep] = true
		}
	}
	newDeps := []string{}
	for _, dep := range currDeps {
		newDeps = append(newDeps, getRequiredBy(db, dep, seen)...)
	}
	return append(currDeps, newDeps...)
}
