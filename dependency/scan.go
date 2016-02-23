package dependency

import (
	"strings"

	"github.com/Masterminds/glide/msg"
	"github.com/Masterminds/glide/util"
)

var osList []string
var archList []string

func init() {
	// The supported systems are listed in
	// https://github.com/golang/go/blob/master/src/go/build/syslist.go
	// The lists are not exported so we need to duplicate them here.
	osListString := "android darwin dragonfly freebsd linux nacl netbsd openbsd plan9 solaris windows"
	osList = strings.Split(osListString, " ")

	archListString := "386 amd64 amd64p32 arm armbe arm64 arm64be ppc64 ppc64le mips mipsle mips64 mips64le mips64p32 mips64p32le ppc s390 s390x sparc sparc64"
	archList = strings.Split(archListString, " ")
}

// IterativeScan attempts to obtain a list of imported dependencies from a
// package. This scanning is different from ImportDir as part of the go/build
// package. It looks over different permutations of the supported OS/Arch to
// try and find all imports. This is different from setting UseAllFiles to
// true on the build Context. It scopes down to just the supported OS/Arch.
//
// Note, there are cases where multiple packages are in the same directory. This
// usually happens with an example that has a main package and a +build tag
// of ignore. This is a bit of a hack. It causes UseAllFiles to have errors.
func IterativeScan(path string) ([]string, error) {

	var pkgs []string
	for _, o := range osList {
		for _, a := range archList {
			b, err := util.GetBuildContext()
			if err != nil {
				return []string{}, err
			}

			// Make sure use all files is off
			b.UseAllFiles = false

			// Set the OS and Arch for this pass
			b.GOARCH = a
			b.GOOS = o

			pk, err := b.ImportDir(path, 0)
			if err != nil {
				msg.Debug("Problem parsing package at %s for %s %s", path, o, a)
				return []string{}, err
			}

			for _, dep := range pk.Imports {
				found := false
				for _, p := range pkgs {
					if p == dep {
						found = true
					}
				}
				if !found {
					pkgs = append(pkgs, dep)
				}
			}
		}
	}

	return pkgs, nil
}
