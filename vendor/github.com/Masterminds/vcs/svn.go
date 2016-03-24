package vcs

import (
	"encoding/xml"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var svnDetectURL = regexp.MustCompile("URL: (?P<foo>.+)\n")

// NewSvnRepo creates a new instance of SvnRepo. The remote and local directories
// need to be passed in. The remote location should include the branch for SVN.
// For example, if the package is https://github.com/Masterminds/cookoo/ the remote
// should be https://github.com/Masterminds/cookoo/trunk for the trunk branch.
func NewSvnRepo(remote, local string) (*SvnRepo, error) {
	ltype, err := DetectVcsFromFS(local)

	// Found a VCS other than Svn. Need to report an error.
	if err == nil && ltype != Svn {
		return nil, ErrWrongVCS
	}

	r := &SvnRepo{}
	r.setRemote(remote)
	r.setLocalPath(local)
	r.Logger = Logger

	// Make sure the local SVN repo is configured the same as the remote when
	// A remote value was passed in.
	if err == nil && r.CheckLocal() == true {
		// An SVN repo was found so test that the URL there matches
		// the repo passed in here.
		out, err := exec.Command("svn", "info", local).CombinedOutput()
		if err != nil {
			return nil, err
		}

		m := svnDetectURL.FindStringSubmatch(string(out))
		if m[1] != "" && m[1] != remote {
			return nil, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Svn repo use that one.
		if remote == "" && m[1] != "" {
			r.setRemote(m[1])
		}
	}

	return r, nil
}

// SvnRepo implements the Repo interface for the Svn source control.
type SvnRepo struct {
	base
}

// Vcs retrieves the underlying VCS being implemented.
func (s SvnRepo) Vcs() Type {
	return Svn
}

// Get is used to perform an initial checkout of a repository.
// Note, because SVN isn't distributed this is a checkout without
// a clone.
func (s *SvnRepo) Get() error {
	_, err := s.run("svn", "checkout", s.Remote(), s.LocalPath())
	return err
}

// Update performs an SVN update to an existing checkout.
func (s *SvnRepo) Update() error {
	_, err := s.runFromDir("svn", "update")
	return err
}

// UpdateVersion sets the version of a package currently checked out via SVN.
func (s *SvnRepo) UpdateVersion(version string) error {
	_, err := s.runFromDir("svn", "update", "-r", version)
	return err
}

// Version retrieves the current version.
func (s *SvnRepo) Version() (string, error) {
	out, err := s.runFromDir("svnversion", ".")
	s.log(out)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Date retrieves the date on the latest commit.
func (s *SvnRepo) Date() (time.Time, error) {
	version, err := s.Version()
	if err != nil {
		return time.Time{}, err
	}
	out, err := s.runFromDir("svn", "pget", "svn:date", "--revprop", "-r", version)
	if err != nil {
		return time.Time{}, err
	}
	const longForm = "2006-01-02T15:04:05.000000Z\n"
	t, err := time.Parse(longForm, string(out))
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// CheckLocal verifies the local location is an SVN repo.
func (s *SvnRepo) CheckLocal() bool {
	if _, err := os.Stat(s.LocalPath() + "/.svn"); err == nil {
		return true
	}

	return false

}

// Tags returns []string{} as there are no formal tags in SVN. Tags are a
// convention in SVN. They are typically implemented as a copy of the trunk and
// placed in the /tags/[tag name] directory. Since this is a convention the
// expectation is to checkout a tag the correct subdirectory will be used
// as the path. For more information see:
// http://svnbook.red-bean.com/en/1.7/svn.branchmerge.tags.html
func (s *SvnRepo) Tags() ([]string, error) {
	return []string{}, nil
}

// Branches returns []string{} as there are no formal branches in SVN. Branches
// are a convention. They are typically implemented as a copy of the trunk and
// placed in the /branches/[tag name] directory. Since this is a convention the
// expectation is to checkout a branch the correct subdirectory will be used
// as the path. For more information see:
// http://svnbook.red-bean.com/en/1.7/svn.branchmerge.using.html
func (s *SvnRepo) Branches() ([]string, error) {
	return []string{}, nil
}

// IsReference returns if a string is a reference. A reference is a commit id.
// Branches and tags are part of the path.
func (s *SvnRepo) IsReference(r string) bool {
	out, err := s.runFromDir("svn", "log", "-r", r)

	// This is a complete hack. There must be a better way to do this. Pull
	// requests welcome. When the reference isn't real you get a line of
	// repeated - followed by an empty line. If the reference is real there
	// is commit information in addition to those. So, we look for responses
	// over 2 lines long.
	lines := strings.Split(string(out), "\n")
	if err == nil && len(lines) > 2 {
		return true
	}

	return false
}

// IsDirty returns if the checkout has been modified from the checked
// out reference.
func (s *SvnRepo) IsDirty() bool {
	out, err := s.runFromDir("svn", "diff")
	return err != nil || len(out) != 0
}

// CommitInfo retrieves metadata about a commit.
func (s *SvnRepo) CommitInfo(id string) (*CommitInfo, error) {
	out, err := s.runFromDir("svn", "log", "-r", id, "--xml")
	if err != nil {
		return nil, err
	}

	type Logentry struct {
		Author string `xml:"author"`
		Date   string `xml:"date"`
		Msg    string `xml:"msg"`
	}
	type Log struct {
		XMLName xml.Name   `xml:"log"`
		Logs    []Logentry `xml:"logentry"`
	}

	logs := &Log{}
	err = xml.Unmarshal(out, &logs)
	if err != nil {
		return nil, err
	}
	if len(logs.Logs) == 0 {
		return nil, ErrRevisionUnavailable
	}

	ci := &CommitInfo{
		Commit:  id,
		Author:  logs.Logs[0].Author,
		Message: logs.Logs[0].Msg,
	}

	if len(logs.Logs[0].Date) > 0 {
		ci.Date, err = time.Parse(time.RFC3339Nano, logs.Logs[0].Date)
		if err != nil {
			return nil, err
		}
	}

	return ci, nil
}
