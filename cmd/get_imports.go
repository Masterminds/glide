package cmd

import (
	"encoding/xml"
	"fmt"
	//"log"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/Masterminds/cookoo"
	v "github.com/Masterminds/vcs"
)

func init() {
	// Precompile the regular expressions used to check VCS locations.
	for _, v := range vcsList {
		v.regex = regexp.MustCompile(v.pattern)
	}

	// Uncomment the line below and the log import to see the output
	// from the vcs commands executed for each project.
	//v.Logger = log.New(os.Stdout, "go-vcs", log.LstdFlags)
}

// GetAll gets zero or more repos.
//
// Params:
//	- packages ([]string): Package names to get.
// 	- verbose (bool): default false
//
// Returns:
// 	- []*Dependency: A list of constructed dependencies.
func GetAll(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	names := p.Get("packages", []string{}).([]string)
	cfg := p.Get("conf", nil).(*Config)
	insecure := p.Get("insecure", false).(bool)

	Info("Preparing to install %d package.", len(names))

	deps := []*Dependency{}
	for _, name := range names {
		cwd, err := VendorPath(c)
		if err != nil {
			return nil, err
		}

		root := getRepoRootFromPackage(name)
		if len(root) == 0 {
			return nil, fmt.Errorf("Package name is required for %q.", name)
		}

		if cfg.HasDependency(root) {
			Warn("Package %q is already in glide.yaml. Skipping", root)
			continue
		}

		dest := path.Join(cwd, root)

		var repoURL string
		if insecure {
			repoURL = "http://" + root
		} else {
			repoURL = "https://" + root
		}
		repo, err := v.NewRepo(repoURL, dest)
		if err != nil {
			Error("Could not construct repo for %q: %s", name, err)
			return false, err
		}

		dep := &Dependency{
			Name: root,
		}

		// When retriving from an insecure location set the repo to the
		// insecure location.
		if insecure {
			dep.Repository = "http://" + root
		}

		subpkg := strings.TrimPrefix(name, root)
		if len(subpkg) > 0 && subpkg != "/" {
			dep.Subpackages = []string{subpkg}
		}

		if err := repo.Get(); err != nil {
			return dep, err
		}

		cfg.Imports = append(cfg.Imports, dep)

		deps = append(deps, dep)

	}
	return deps, nil
}

// GetImports iterates over the imported packages and gets them.
func GetImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)
	cwd, err := VendorPath(c)
	if err != nil {
		Error("Failed to prepare vendor directory: %s", err)
		return false, err
	}

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing downloaded.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsGet(dep, cwd); err != nil {
			Warn("Skipped getting %s: %v\n", dep.Name, err)
		}
	}

	return true, nil
}

// UpdateImports iterates over the imported packages and updates them.
//
// Params:
//
// 	- force (bool): force packages to update (default false)
//	- conf (*Config): The configuration
// 	- packages([]string): The packages to update. Default is all.
func UpdateImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)
	force := p.Get("force", true).(bool)
	plist := p.Get("packages", []string{}).([]string)
	pkgs := list2map(plist)
	restrict := len(pkgs) > 0

	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing updated.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if restrict && !pkgs[dep.Name] {
			Debug("===> Skipping %q", dep.Name)
			continue
		}
		if err := VcsUpdate(dep, cwd, force); err != nil {
			Warn("Update failed for %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

// SetReference is a command to set the VCS reference (commit id, tag, etc) for
// a project.
func SetReference(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)
	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(cfg.Imports) == 0 {
		Info("No references set.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsVersion(dep, cwd); err != nil {
			Warn("Failed to set version on %s to %s: %s\n", dep.Name, dep.Reference, err)
		}
	}

	return true, nil
}

// filterArchOs indicates a dependency should be filtered out because it is
// the wrong GOOS or GOARCH.
func filterArchOs(dep *Dependency) bool {
	found := false
	if len(dep.Arch) > 0 {
		for _, a := range dep.Arch {
			if a == runtime.GOARCH {
				found = true
			}
		}
		// If it's not found, it should be filtered out.
		if !found {
			return true
		}
	}

	found = false
	if len(dep.Os) > 0 {
		for _, o := range dep.Os {
			if o == runtime.GOOS {
				found = true
			}
		}
		if !found {
			return true
		}

	}

	return false
}

// VcsExists checks if the directory has a local VCS checkout.
func VcsExists(dep *Dependency, dest string) bool {
	repo, err := dep.GetRepo(dest)
	if err != nil {
		return false
	}

	return repo.CheckLocal()
}

// VcsGet figures out how to fetch a dependency, and then gets it.
//
// VcsGet installs into the dest.
func VcsGet(dep *Dependency, dest string) error {

	repo, err := dep.GetRepo(dest)
	if err != nil {
		return err
	}

	return repo.Get()
}

// VcsUpdate updates to a particular checkout based on the VCS setting.
func VcsUpdate(dep *Dependency, vend string, force bool) error {
	Info("Fetching updates for %s.\n", dep.Name)

	if filterArchOs(dep) {
		Info("%s is not used for %s/%s.\n", dep.Name, runtime.GOOS, runtime.GOARCH)
		return nil
	}

	dest := path.Join(vend, dep.Name)
	// If destination doesn't exist we need to perform an initial checkout.
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err = VcsGet(dep, dest); err != nil {
			Warn("Unable to checkout %s\n", dep.Name)
			return err
		}
	} else {
		// At this point we have a directory for the package.

		// When the directory is not empty and has no VCS directory it's
		// a vendored files situation.
		empty, err := isDirectoryEmpty(dest)
		if err != nil {
			return err
		}
		_, err = v.DetectVcsFromFS(dest)
		if empty == false && err == v.ErrCannotDetectVCS {
			Warn("%s appears to be a vendored package. Unable to update. Consider the '--update-vendored' flag.\n", dep.Name)
		} else {
			repo, err := dep.GetRepo(dest)

			// Tried to checkout a repo to a path that does not work. Either the
			// type or endpoint has changed. Force is being passed in so the old
			// location can be removed and replaced with the new one.
			// Warning, any changes in the old location will be deleted.
			// TODO: Put dirty checking in on the existing local checkout.
			if (err == v.ErrWrongVCS || err == v.ErrWrongRemote) && force == true {
				var newRemote string
				if len(dep.Repository) > 0 {
					newRemote = dep.Repository
				} else {
					newRemote = "https://" + dep.Name
				}

				Warn("Replacing %s with contents from %s\n", dep.Name, newRemote)
				rerr := os.RemoveAll(dest)
				if rerr != nil {
					return rerr
				}
				if err = VcsGet(dep, dest); err != nil {
					Warn("Unable to checkout %s\n", dep.Name)
					return err
				}
			} else if err != nil {
				return err
			} else {
				if err := repo.Update(); err != nil {
					Warn("Download failed.\n")
					return err
				}
			}
		}
	}

	return nil
}

// VcsVersion set the VCS version for a checkout.
func VcsVersion(dep *Dependency, vend string) error {
	// If there is no refernece configured there is nothing to set.
	if dep.Reference == "" {
		return nil
	}

	cwd := path.Join(vend, dep.Name)

	// When the directory is not empty and has no VCS directory it's
	// a vendored files situation.
	empty, err := isDirectoryEmpty(cwd)
	if err != nil {
		return err
	}
	_, err = v.DetectVcsFromFS(cwd)
	if empty == false && err == v.ErrCannotDetectVCS {
		Warn("%s appears to be a vendored package. Unable to set new version. Consider the '--update-vendored' flag.\n", dep.Name)
	} else {

		Info("Setting version for %s.\n", dep.Name)

		repo, err := dep.GetRepo(cwd)
		if err != nil {
			return err
		}

		if err := repo.UpdateVersion(dep.Reference); err != nil {
			Error("Failed to set version to %s: %s\n", dep.Reference, err)
			return err
		}
	}

	return nil
}

// VcsLastCommit gets the last commit ID from the given dependency.
func VcsLastCommit(dep *Dependency, vend string) (string, error) {
	cwd := path.Join(vend, dep.Name)
	repo, err := dep.GetRepo(cwd)
	if err != nil {
		return "", err
	}

	if repo.CheckLocal() == false {
		return "", fmt.Errorf("%s is not a VCS repo\n", dep.Name)
	}

	version, err := repo.Version()
	if err != nil {
		return "", err
	}

	return version, nil
}

// From a package name find the root repo. For example,
// the package github.com/Masterminds/cookoo/io has a root repo
// at github.com/Masterminds/cookoo
func getRepoRootFromPackage(pkg string) string {
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
	pkg = getRepoRootFromGoGet(pkg)

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
func getRepoRootFromGoGet(pkg string) string {

	vcsURL := "https://" + pkg
	u, err := url.Parse(vcsURL)
	if err != nil {
		return pkg
	}
	if u.RawQuery == "" {
		u.RawQuery = "go-get=1"
	} else {
		u.RawQuery = u.RawQuery + "+go-get=1"
	}
	checkURL := u.String()
	resp, err := http.Get(checkURL)
	if err != nil {
		return pkg
	}
	defer resp.Body.Close()

	nu, err := parseImportFromBody(u, resp.Body)
	if err != nil {
		return pkg
	} else if nu == "" {
		return pkg
	}

	return nu
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
				err = nil
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

			// If this the second time a go-import statement has been detected
			// return an error. There should only be one import statement per
			// html file. We don't simply return the first found in order to
			// detect pages including more than one.
			if u != "" {
				u = ""
				err = v.ErrCannotDetectVCS
				return
			}

			// If the prefix supplied by the remote system isn't a prefix to the
			// url we're fetching return an error. This will work for exact
			// matches and prefixes. For example, golang.org/x/net as a prefix
			// will match for golang.org/x/net and golang.org/x/net/context.
			vcsURL := ur.Host + ur.Path
			if !strings.HasPrefix(vcsURL, f[0]) {
				err = v.ErrCannotDetectVCS
				return
			}

			u = f[0]
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
