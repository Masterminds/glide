package repo

import "testing"

func TestUpdateTracker(t *testing.T) {
	tr := NewUpdateTracker()

	if f := tr.Check("github.com/foo/bar"); f != false {
		t.Error("Error, package Check passed on empty tracker")
	}

	tr.Add("github.com/foo/bar")

	if f := tr.Check("github.com/foo/bar"); f != true {
		t.Error("Error, failed to add package to tracker")
	}

	tr.Remove("github.com/foo/bar")

	if f := tr.Check("github.com/foo/bar"); f != false {
		t.Error("Error, failed to remove package from tracker")
	}
}
