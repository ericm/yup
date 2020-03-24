package search

import (
	"testing"
)

func TestAur(t *testing.T) {
	shouldBe, err := Aur("test", false, true)
	if err != nil {
		t.Error(err)
	}
	if len(shouldBe) == 0 {
		t.Error("yup test returned no packages")
	}
}
