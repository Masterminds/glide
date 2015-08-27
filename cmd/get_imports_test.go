package cmd

import (
	"testing"

	"github.com/Masterminds/cookoo"
)

func TestGetImportsEmptyConfig(t *testing.T) {
	_, _, c := cookoo.Cookoo()
	SilenceLogs(c)
	cfg := new(Config)
	p := cookoo.NewParamsWithValues(map[string]interface{}{"conf": cfg})
	res, it := GetImports(c, p)
	if it != nil {
		t.Errorf("Interrupt value non-nil")
	}
	bres, ok := res.(bool)
	if !ok || bres {
		t.Errorf("Result was non-bool or true: ok=%t bres=%t", ok, bres)
	}
}

func SilenceLogs(c cookoo.Context) {
	p := cookoo.NewParamsWithValues(map[string]interface{}{"quiet": true})
	BeQuiet(c, p)
}

func TestGetRepoRootFromPackage(t *testing.T) {
	urlList := map[string]string{
		"github.com/Masterminds/VCSTestRepo":                       "github.com/Masterminds/VCSTestRepo",
		"bitbucket.org/mattfarina/testhgrepo":                      "bitbucket.org/mattfarina/testhgrepo",
		"launchpad.net/govcstestbzrrepo/trunk":                     "launchpad.net/govcstestbzrrepo/trunk",
		"launchpad.net/~mattfarina/+junk/mygovcstestbzrrepo":       "launchpad.net/~mattfarina/+junk/mygovcstestbzrrepo",
		"launchpad.net/~mattfarina/+junk/mygovcstestbzrrepo/trunk": "launchpad.net/~mattfarina/+junk/mygovcstestbzrrepo",
		"git.launchpad.net/govcstestgitrepo":                       "git.launchpad.net/govcstestgitrepo",
		"git.launchpad.net/~mattfarina/+git/mygovcstestgitrepo":    "git.launchpad.net/~mattfarina/+git/mygovcstestgitrepo",
		"farbtastic.googlecode.com/svn/":                           "farbtastic.googlecode.com/svn/",
		"farbtastic.googlecode.com/svn/trunk":                      "farbtastic.googlecode.com/svn/trunk",
		"code.google.com/p/farbtastic":                             "code.google.com/p/farbtastic",
		"code.google.com/p/plotinum":                               "code.google.com/p/plotinum",
		"example.com/foo/bar.git":                                  "example.com/foo/bar.git",
		"example.com/foo/bar.svn":                                  "example.com/foo/bar.svn",
		"example.com/foo/bar/baz.bzr":                              "example.com/foo/bar/baz.bzr",
		"example.com/foo/bar/baz.hg":                               "example.com/foo/bar/baz.hg",
		"gopkg.in/mgo.v2":                                          "gopkg.in/mgo.v2",
		"gopkg.in/mgo.v2/txn":                                      "gopkg.in/mgo.v2",
		"gopkg.in/nowk/assert.v2":                                  "gopkg.in/nowk/assert.v2",
		"gopkg.in/nowk/assert.v2/tests":                            "gopkg.in/nowk/assert.v2",
		"golang.org/x/net":                                         "golang.org/x/net",
		"golang.org/x/net/context":                                 "golang.org/x/net",
	}

	for u, c := range urlList {
		repo := getRepoRootFromPackage(u)
		if repo != c {
			t.Errorf("getRepoRootFromPackage expected %s but got %s", c, repo)
		}
	}
}
