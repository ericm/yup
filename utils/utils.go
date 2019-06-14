package utils

import (
	"github.com/ericm/yup/sync"
	"strconv"
	"strings"
)

// ParseNumbers adds packages to packsToInstall
func ParseNumbers(input string, packs []sync.PkgBuild, packsToInstall *[]sync.PkgBuild) {
	inputs := strings.Split((strings.ToLower(strings.TrimSpace(input))), " ")
	seen := map[int]bool{}
	for _, s := range inputs {
		// 1-3
		if strings.Contains(s, "-") {
			continue
		}
		// ^4
		if strings.Contains(s, "^") {
			continue
		}

		if num, err := strconv.Atoi(s); err == nil {
			// Find package from input
			index := len(packs) - num
			// Add to the slice
			if index < len(packs) && index >= 0 && !seen[index] {
				*packsToInstall = append(*packsToInstall, packs[index])
				seen[index] = true
			}
		}
	}
}
