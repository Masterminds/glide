package cmd

// VCS provides the interface to work with different source control systems such
// as Git, Bzr, Mercurial, and SVN. For implementations of this interface see
// BzrVCS, GitVCS, HgVCS, and SvnVCS.
type VCS interface {

	// Get is used to perform an initial checkout of a repository.
	Get(*Dependency) error

	// Update performs an update to an existing checkout of a repository.
	Update(*Dependency) error

	// Version sets the version of a package of a repository.
	Version(*Dependency) error

	// LastCommit retrieves the current version.
	LastCommit(*Dependency) (string, error)
}
