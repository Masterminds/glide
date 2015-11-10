package cmd

import (
	"container/list"
	"testing"
)

func TestFindInTree(t *testing.T) {
	l := list.New()
	l.PushBack("github.com/Masterminds/glide")
	l.PushBack("github.com/Masterminds/vcs")
	l.PushBack("github.com/Masterminds/semver")

	f := findInList("foo", l)
	if f != false {
		t.Error("findInList found true instead of false")
	}

	f = findInList("github.com/Masterminds/vcs", l)
	if f != true {
		t.Error("findInList found false instead of true")
	}
}
