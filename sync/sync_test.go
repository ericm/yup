package sync

import (
	"testing"
)

func TestParseNumbers(t *testing.T) {
	packs := []PkgBuild{PkgBuild{}, PkgBuild{}, PkgBuild{}, PkgBuild{}}
	ParseNumbers("1 2 3", &packs)
	if len(packs) != 3 {
		t.Errorf("Filter for '1 2 3' returned %d results", len(packs))
	}
}
