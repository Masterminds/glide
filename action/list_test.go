package action

import "testing"

func TestList(t *testing.T) {
	if len(List("../", false).Installed) < 1 {
		t.Error("Expected some packages to be found")
	}
}
