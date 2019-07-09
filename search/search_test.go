package search

import (
	"testing"
)

func TestAur(t *testing.T) {
	should_be, err := Aur("test", false, true)
	if err != nil {
		t.Error(err)
	}
	if len(should_be) == 0 {
		t.Error("yup test returned no packages")
	}
}
