package action

import (
	"container/list"
	"os"

	"github.com/Masterminds/glide/msg"
	"github.com/Masterminds/glide/tree"
	"github.com/Masterminds/glide/util"
)

// Tree prints a tree representing dependencies.
func Tree(basedir string, showcore bool) {
	msg.Warn("The tree command is deprecated and will be removed in a future version")
	buildContext, err := util.GetBuildContext()
	if err != nil {
		msg.Die("Failed to get a build context: %s", err)
	}
	myName := buildContext.PackageName(basedir)

	if basedir == "." {
		var err error
		basedir, err = os.Getwd()
		if err != nil {
			msg.Die("Could not get working directory")
		}
	}

	msg.Puts(myName)
	l := list.New()
	l.PushBack(myName)
	tree.Display(buildContext, basedir, myName, 1, showcore, l)
}
