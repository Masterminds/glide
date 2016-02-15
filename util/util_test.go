package util

import "testing"

func TestGetRootFromPackage(t *testing.T) {
	urlList := map[string]string{
		"github.com/Masterminds/VCSTestRepo":                       "github.com/Masterminds/VCSTestRepo",
		"bitbucket.org/mattfarina/testhgrepo":                      "bitbucket.org/mattfarina/testhgrepo",
		"launchpad.net/govcstestbzrrepo/trunk":                     "launchpad.net/govcstestbzrrepo/trunk",
		"launchpad.net/~mattfarina/+junk/mygovcstestbzrrepo":       "launchpad.net/~mattfarina/+junk/mygovcstestbzrrepo",
		"launchpad.net/~mattfarina/+junk/mygovcstestbzrrepo/trunk": "launchpad.net/~mattfarina/+junk/mygovcstestbzrrepo",
		"git.launchpad.net/govcstestgitrepo":                       "git.launchpad.net/govcstestgitrepo",
		"git.launchpad.net/~mattfarina/+git/mygovcstestgitrepo":    "git.launchpad.net/~mattfarina/+git/mygovcstestgitrepo",
		"hub.jazz.net/git/user/pkgname":                            "hub.jazz.net/git/user/pkgname",
		"hub.jazz.net/git/user/pkgname/subpkg/subpkg/subpkg":       "hub.jazz.net/git/user/pkgname",
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
		repo := GetRootFromPackage(u)
		if repo != c {
			t.Errorf("getRepoRootFromPackage expected %s but got %s", c, repo)
		}
	}
}
