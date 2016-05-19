package util

import (
	"encoding/xml"
	"fmt"
	"go/build"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/vcs"
)

// ResolveCurrent selects whether the package should only the dependencies for
// the current OS/ARCH instead of all possible permutations.
// This is not concurrently safe which is ok for the current application. If
// other needs arise it may need to be re-written.
var ResolveCurrent = false

func init() {
	// Precompile the regular expressions used to check VCS locations.
	for _, v := range vcsList {
		v.regex = regexp.MustCompile(v.pattern)
	}
}

func toSlash(v string) string {
	return strings.Replace(v, "\\", "/", -1)
}

// GetRootFromPackage retrives the top level package from a name.
//
// From a package name find the root repo. For example,
// the package github.com/Masterminds/cookoo/io has a root repo
// at github.com/Masterminds/cookoo
func GetRootFromPackage(pkg string) string {
	pkg = toSlash(pkg)
	for _, v := range vcsList {
		m := v.regex.FindStringSubmatch(pkg)
		if m == nil {
			continue
		}

		if m[1] != "" {
			return m[1]
		}
	}

	// There are cases where a package uses the special go get magic for
	// redirects. If we've not discovered the location already try that.
	pkg = getRootFromGoGet(pkg)

	return pkg
}

// Pages like https://golang.org/x/net provide an html document with
// meta tags containing a location to work with. The go tool uses
// a meta tag with the name go-import which is what we use here.
// godoc.org also has one call go-source that we do not need to use.
// The value of go-import is in the form "prefix vcs repo". The prefix
// should match the vcsURL and the repo is a location that can be
// checked out. Note, to get the html document you you need to add
// ?go-get=1 to the url.
func getRootFromGoGet(pkg string) string {

	p, found := checkRemotePackageCache(pkg)
	if found {
		return p
	}

	vcsURL := "https://" + pkg
	u, err := url.Parse(vcsURL)
	if err != nil {
		return pkg
	}
	if u.RawQuery == "" {
		u.RawQuery = "go-get=1"
	} else {
		u.RawQuery = u.RawQuery + "&go-get=1"
	}
	checkURL := u.String()
	resp, err := http.Get(checkURL)
	if err != nil {
		addToRemotePackageCache(pkg, pkg)
		return pkg
	}
	defer resp.Body.Close()

	nu, err := parseImportFromBody(u, resp.Body)
	if err != nil {
		addToRemotePackageCache(pkg, pkg)
		return pkg
	} else if nu == "" {
		addToRemotePackageCache(pkg, pkg)
		return pkg
	}

	addToRemotePackageCache(pkg, nu)
	return nu
}

// The caching is not concurrency safe but should be made to be that way.
// This implementation is far too much of a hack... rewrite needed.
var remotePackageCache = make(map[string]string)

func checkRemotePackageCache(pkg string) (string, bool) {
	for k, v := range remotePackageCache {
		if pkg == k {
			return v, true
		}
	}

	return pkg, false
}

func addToRemotePackageCache(pkg, v string) {
	remotePackageCache[pkg] = v
}

func parseImportFromBody(ur *url.URL, r io.ReadCloser) (u string, err error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false
	var t xml.Token
	for {
		t, err = d.Token()
		if err != nil {
			if err == io.EOF {
				// If we hit the end of the markup and don't have anything
				// we return an error.
				err = vcs.ErrCannotDetectVCS
			}
			return
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "body") {
			return
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, "head") {
			return
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, "meta") {
			continue
		}
		if attrValue(e.Attr, "name") != "go-import" {
			continue
		}
		if f := strings.Fields(attrValue(e.Attr, "content")); len(f) == 3 {

			// If the prefix supplied by the remote system isn't a prefix to the
			// url we're fetching return continue looking for more go-imports.
			// This will work for exact matches and prefixes. For example,
			// golang.org/x/net as a prefix will match for golang.org/x/net and
			// golang.org/x/net/context.
			vcsURL := ur.Host + ur.Path
			if !strings.HasPrefix(vcsURL, f[0]) {
				continue
			} else {
				u = f[0]
				return
			}

		}
	}
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("can't decode XML document using charset %q", charset)
	}
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}

type vcsInfo struct {
	host    string
	pattern string
	regex   *regexp.Regexp
}

var vcsList = []*vcsInfo{
	{
		host:    "github.com",
		pattern: `^(?P<rootpkg>github\.com/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
	},
	{
		host:    "bitbucket.org",
		pattern: `^(?P<rootpkg>bitbucket\.org/([A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+))(/[A-Za-z0-9_.\-]+)*$`,
	},
	{
		host:    "launchpad.net",
		pattern: `^(?P<rootpkg>launchpad\.net/(([A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)?|~[A-Za-z0-9_.\-]+/(\+junk|[A-Za-z0-9_.\-]+)/[A-Za-z0-9_.\-]+))(/[A-Za-z0-9_.\-]+)*$`,
	},
	{
		host:    "git.launchpad.net",
		pattern: `^(?P<rootpkg>git\.launchpad\.net/(([A-Za-z0-9_.\-]+)|~[A-Za-z0-9_.\-]+/(\+git|[A-Za-z0-9_.\-]+)/[A-Za-z0-9_.\-]+))$`,
	},
	{
		host:    "hub.jazz.net",
		pattern: `^(?P<rootpkg>hub\.jazz\.net/git/[a-z0-9]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
	},
	{
		host:    "go.googlesource.com",
		pattern: `^(?P<rootpkg>go\.googlesource\.com/[A-Za-z0-9_.\-]+/?)$`,
	},
	// TODO: Once Google Code becomes fully deprecated this can be removed.
	{
		host:    "code.google.com",
		pattern: `^(?P<rootpkg>code\.google\.com/[pr]/([a-z0-9\-]+)(\.([a-z0-9\-]+))?)(/[A-Za-z0-9_.\-]+)*$`,
	},
	// Alternative Google setup for SVN. This is the previous structure but it still works... until Google Code goes away.
	{
		pattern: `^(?P<rootpkg>[a-z0-9_\-.]+\.googlecode\.com/svn(/.*)?)$`,
	},
	// Alternative Google setup. This is the previous structure but it still works... until Google Code goes away.
	{
		pattern: `^(?P<rootpkg>[a-z0-9_\-.]+\.googlecode\.com/(git|hg))(/.*)?$`,
	},
	// If none of the previous detect the type they will fall to this looking for the type in a generic sense
	// by the extension to the path.
	{
		pattern: `^(?P<rootpkg>(?P<repo>([a-z0-9.\-]+\.)+[a-z0-9.\-]+(:[0-9]+)?/[A-Za-z0-9_.\-/]*?)\.(bzr|git|hg|svn))(/[A-Za-z0-9_.\-]+)*$`,
	},
}

// BuildCtxt is a convenience wrapper for not having to import go/build
// anywhere else
type BuildCtxt struct {
	build.Context
}

// PackageName attempts to determine the name of the base package.
//
// If resolution fails, this will return "main".
func (b *BuildCtxt) PackageName(base string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return "main"
	}

	pkg, err := b.Import(base, cwd, 0)
	if err != nil {
		// There may not be any top level Go source files but the project may
		// still be within the GOPATH.
		if strings.HasPrefix(base, b.GOPATH) {
			p := strings.TrimPrefix(base, filepath.Join(b.GOPATH, "src"))
			return strings.Trim(p, string(os.PathSeparator))
		}
	}

	return pkg.ImportPath
}

// GetBuildContext returns a build context from go/build. When the $GOROOT
// variable is not set in the users environment it sets the context's root
// path to the path returned by 'go env GOROOT'.
//
// TODO: This should be moved to the `dependency` package.
func GetBuildContext() (*BuildCtxt, error) {
	buildContext := &BuildCtxt{build.Default}

	// If we aren't resolving for the current system set to look at all
	// build modes.
	if !ResolveCurrent {
		// This tells the context scanning to skip filtering on +build flags or
		// file names.
		buildContext.UseAllFiles = true
	}

	if goRoot := os.Getenv("GOROOT"); len(goRoot) == 0 {
		goExecutable := os.Getenv("GLIDE_GO_EXECUTABLE")
		if len(goExecutable) <= 0 {
			goExecutable = "go"
		}
		out, err := exec.Command(goExecutable, "env", "GOROOT").Output()
		if goRoot = strings.TrimSpace(string(out)); len(goRoot) == 0 || err != nil {
			return nil, fmt.Errorf("Please set the $GOROOT environment " +
				"variable to use this command\n")
		}
		buildContext.GOROOT = goRoot
	}
	return buildContext, nil
}

// NormalizeName takes a package name and normalizes it to the top level package.
//
// For example, golang.org/x/crypto/ssh becomes golang.org/x/crypto. 'ssh' is
// returned as extra data.
//
// FIXME: Is this deprecated?
func NormalizeName(name string) (string, string) {
	// Fastpath check if a name in the GOROOT. There is an issue when a pkg
	// is in the GOROOT and GetRootFromPackage tries to look it up because it
	// expects remote names.
	b, err := GetBuildContext()
	if err == nil {
		p := filepath.Join(b.GOROOT, "src", name)
		if _, err := os.Stat(p); err == nil {
			return toSlash(name), ""
		}
	}

	name = toSlash(name)
	root := GetRootFromPackage(name)
	extra := strings.TrimPrefix(name, root)
	if len(extra) > 0 && extra != "/" {
		extra = strings.TrimPrefix(extra, "/")
	} else {
		// If extra is / (which is what it would be here) we want to return ""
		extra = ""
	}

	return root, extra
}
