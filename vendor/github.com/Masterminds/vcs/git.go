package vcs

import (
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NewGitRepo creates a new instance of GitRepo. The remote and local directories
// need to be passed in.
func NewGitRepo(remote, local string) (*GitRepo, error) {
	ltype, err := DetectVcsFromFS(local)

	// Found a VCS other than Git. Need to report an error.
	if err == nil && ltype != Git {
		return nil, ErrWrongVCS
	}

	r := &GitRepo{}
	r.setRemote(remote)
	r.setLocalPath(local)
	r.RemoteLocation = "origin"
	r.Logger = Logger

	// Make sure the local Git repo is configured the same as the remote when
	// A remote value was passed in.
	if err == nil && r.CheckLocal() == true {
		c := exec.Command("git", "config", "--get", "remote.origin.url")
		c.Dir = local
		c.Env = envForDir(c.Dir)
		out, err := c.CombinedOutput()
		if err != nil {
			return nil, err
		}

		localRemote := strings.TrimSpace(string(out))
		if remote != "" && localRemote != remote {
			return nil, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Git repo use that one.
		if remote == "" && localRemote != "" {
			r.setRemote(localRemote)
		}
	}

	return r, nil
}

// GitRepo implements the Repo interface for the Git source control.
type GitRepo struct {
	base
	RemoteLocation string
}

// Vcs retrieves the underlying VCS being implemented.
func (s GitRepo) Vcs() Type {
	return Git
}

// Get is used to perform an initial clone of a repository.
func (s *GitRepo) Get() error {
	_, err := s.run("git", "clone", s.Remote(), s.LocalPath())

	// There are some windows cases where Git cannot create the parent directory,
	// if it does not already exist, to the location it's trying to create the
	// repo. Catch that error and try to handle it.
	if err != nil && s.isUnableToCreateDir(err) {

		basePath := filepath.Dir(filepath.FromSlash(s.LocalPath()))
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			err = os.MkdirAll(basePath, 0755)
			if err != nil {
				return err
			}

			_, err = s.run("git", "clone", s.Remote(), s.LocalPath())
			return err
		}

	}

	return err
}

// Update performs an Git fetch and pull to an existing checkout.
func (s *GitRepo) Update() error {
	// Perform a fetch to make sure everything is up to date.
	_, err := s.runFromDir("git", "fetch", s.RemoteLocation)
	if err != nil {
		return err
	}

	// When in a detached head state, such as when an individual commit is checked
	// out do not attempt a pull. It will cause an error.
	detached, err := isDetachedHead(s.LocalPath())

	if err != nil {
		return err
	}

	if detached == true {
		return nil
	}

	_, err = s.runFromDir("git", "pull")
	return err
}

// UpdateVersion sets the version of a package currently checked out via Git.
func (s *GitRepo) UpdateVersion(version string) error {
	_, err := s.runFromDir("git", "checkout", version)
	return err
}

// Version retrieves the current version.
func (s *GitRepo) Version() (string, error) {
	out, err := s.runFromDir("git", "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// Date retrieves the date on the latest commit.
func (s *GitRepo) Date() (time.Time, error) {
	out, err := s.runFromDir("git", "log", "-1", "--date=iso", "--pretty=format:%cd")
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse(longForm, string(out))
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// Branches returns a list of available branches on the RemoteLocation
func (s *GitRepo) Branches() ([]string, error) {
	out, err := s.runFromDir("git", "show-ref")
	if err != nil {
		return []string{}, err
	}
	branches := s.referenceList(string(out), `(?m-s)(?:`+s.RemoteLocation+`)/(\S+)$`)
	return branches, nil
}

// Tags returns a list of available tags on the RemoteLocation
func (s *GitRepo) Tags() ([]string, error) {
	out, err := s.runFromDir("git", "show-ref")
	if err != nil {
		return []string{}, err
	}
	tags := s.referenceList(string(out), `(?m-s)(?:tags)/(\S+)$`)
	return tags, nil
}

// CheckLocal verifies the local location is a Git repo.
func (s *GitRepo) CheckLocal() bool {
	if _, err := os.Stat(s.LocalPath() + "/.git"); err == nil {
		return true
	}

	return false
}

// IsReference returns if a string is a reference. A reference can be a
// commit id, branch, or tag.
func (s *GitRepo) IsReference(r string) bool {
	_, err := s.runFromDir("git", "rev-parse", "--verify", r)
	if err == nil {
		return true
	}

	// Some refs will fail rev-parse. For example, a remote branch that has
	// not been checked out yet. This next step should pickup the other
	// possible references.
	_, err = s.runFromDir("git", "show-ref", r)
	if err == nil {
		return true
	}

	return false
}

// IsDirty returns if the checkout has been modified from the checked
// out reference.
func (s *GitRepo) IsDirty() bool {
	out, err := s.runFromDir("git", "diff")
	return err != nil || len(out) != 0
}

// CommitInfo retrieves metadata about a commit.
func (s *GitRepo) CommitInfo(id string) (*CommitInfo, error) {
	fm := `--pretty=format:"<logentry><commit>%H</commit><author>%an &lt;%ae&gt;</author><date>%aD</date><message>%s</message></logentry>"`
	out, err := s.runFromDir("git", "log", id, fm, "-1")
	if err != nil {
		return nil, ErrRevisionUnavailable
	}

	cis := struct {
		Commit  string `xml:"commit"`
		Author  string `xml:"author"`
		Date    string `xml:"date"`
		Message string `xml:"message"`
	}{}
	err = xml.Unmarshal(out, &cis)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("Mon, _2 Jan 2006 15:04:05 -0700", cis.Date)
	if err != nil {
		return nil, err
	}

	ci := &CommitInfo{
		Commit:  cis.Commit,
		Author:  cis.Author,
		Date:    t,
		Message: cis.Message,
	}

	return ci, nil
}

func isDetachedHead(dir string) (bool, error) {
	c := exec.Command("git", "status", "-uno")
	c.Dir = dir
	c.Env = envForDir(c.Dir)
	out, err := c.CombinedOutput()
	if err != nil {
		return false, err
	}
	detached := strings.Contains(string(out), "HEAD detached at")

	return detached, nil
}

// In a multi-langual manner check for the Git error that it couldn't create
// the directory.
func (s *GitRepo) isUnableToCreateDir(err error) bool {
	msg := err.Error()
	if strings.HasPrefix(msg, "could not create work tree dir") ||
		strings.HasPrefix(msg, "不能创建工作区目录") ||
		strings.HasPrefix(msg, "no s'ha pogut crear el directori d'arbre de treball") ||
		strings.HasPrefix(msg, "impossible de créer le répertoire de la copie de travail") ||
		strings.HasPrefix(msg, "kunde inte skapa arbetskatalogen") ||
		(strings.HasPrefix(msg, "Konnte Arbeitsverzeichnis") && strings.Contains(msg, "nicht erstellen")) ||
		(strings.HasPrefix(msg, "작업 디렉터리를") && strings.Contains(msg, "만들 수 없습니다")) {
		return true
	}

	return false
}
