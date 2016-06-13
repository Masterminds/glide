# 1.5.1 (2016-03-23)

- Fixing bug parsing some Git commit dates.

# 1.5.0 (2016-03-22)

- Add Travis CI testing for Go 1.6.
- Issue #17: Add CommitInfo method allowing for a common way to get commit
  metadata from all VCS.
- Autodetect types that have git@ or hg@ users.
- Autodetect git+ssh, bzr+ssh, git, and svn+ssh scheme urls.
- On Bitbucket for ssh style URLs retrieve the type from the URL. This allows
  for private repo type detection.
- Issue #14: Autodetect ssh/scp style urls (thanks chonthu).

# 1.4.1 (2016-03-07)

- Fixes #16: some windows situations are unable to create parent directory.

# 1.4.0 (2016-02-15)

- Adding support for IBM JazzHub.

# 1.3.1 (2016-01-27)

- Issue #12: Failed to checkout Bzr repo when parent directory didn't
  exist (thanks cyrilleverrier).

# 1.3.0 (2015-11-09)

- Issue #9: Added Date method to get the date/time of latest commit (thanks kamilchm).

# 1.2.0 (2015-10-29)

- Adding IsDirty method to detect a checkout with uncommitted changes.

# 1.1.4 (2015-10-28)

- Fixed #8: Git IsReference not detecting branches that have not been checked
  out yet.

# 1.1.3 (2015-10-21)

- Fixing issue where there are multiple go-import statements for go redirects

# 1.1.2 (2015-10-20)

- Fixes #7: hg not checking out code when Get is called

# 1.1.1 (2015-10-20)

- Issue #6: Allow VCS commands to be run concurrently.

# 1.1.0 (2015-10-19)

- #5: Added output of failed command to returned errors.

# 1.0.0 (2015-10-06)

- Initial release.
