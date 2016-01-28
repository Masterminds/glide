/* Package tree contains functions for printing a dependency tree.

The future of the tree functionality is uncertain, as it is neither core to
the functionality of Glide, nor particularly complementary. Its principal use
case is for debugging the generated dependency tree.

Currently, the tree package builds its dependency tree in a slightly different
way than the `dependency` package does. This should not make any practical
difference, though code-wise it would be nice to change this over to use the
`dependency` resolver.
*/
package tree

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
