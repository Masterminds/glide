# Release 0.6.1 (2015-09-21)

- Fixed #82: C was not recognized as an internal package.
- Fixed #84: novendor (nv) command returned directories with no Go code.

# Release 0.6.0 (2015-09-16)

- #53: Add support for gb-vendor manifest files.
- Added `glide tree` command to inspect the code and see the imported packages.
- Added `glide list` to see an alphabetized list of imported projects.
- Added flatten feature to flatten the vendor tree (thanks interlock).
- Fixed #74: Glide guess using the wrong GOROOT locations in some environments
  (thanks janeczku).
- Fixed #76: Glide tree doesn't exclude core libraries with the GOROOT is
  incorrect (thanks janeczku).
- Fixed #81: rebuild command did not look in vendor/ directory
- Fixed #77: update failed when a commit id was set for the ref

# Release 0.5.1 (2015-08-31)

- Fixed #58: Guess command not working.
- Fixed #56: Unable to use glide get on golang.org/x/[name]/[subpackage]
- Fixed #61: The wrong version of a dependency can be pinned when packages are
  vendored (no VCS repo associated with them).
- Fixed #67: Unable to work go-get redirects.
- Fixed #66: 'glide up' now has an --update-vendored (-u) flag to update
  vendored directories.
- Fixed #68: Handling the base where the GOPATH has multiple separated directories.

# Release 0.5.0 (2015-08-19)

**Glide .5 is a major update breaking some backwards compatability with
previous releases.**

- Migrated to using the vendor/ directory and the go tools for vendor
  package management. To leverage this you'll need to set the
  environment variable GO15VENDOREXPERIMENT=1 and use Go 1.5.
- `glide up` is now recursive and walks installed packages if there is
  no vendor directory. Use the --no-recursive flag to skip this.
- Removed GOPATH management. This was needed for vendor package
  management that's not built into the go toolchain.
- Switched to github.com/Masterminds/vcs for VCS integration.
- When updating packages are now deleted if the --delete flag is set.
  This feature is now opt-in.
- Fixed #32: Detects VCS type and endpoint changes along with a --force flag
  to replace the checkout if desired.

# Release 0.4.1 (2015-07-13)

- Issue #48: When GOPATH not _vendor directory not deleting unused packages.

# Release 0.4.0 (2015-07-07)

- Issue #34: Delete unused packages on update unless flag set.
- Added 'glide create PACKAGE'
- Added 'glide exec COMMAND'
- Added 'glide get PACKAGE'
- Added 'glide pin FILENAME'
- Added 'glide guess FILENAME'
- Updated help text

# Release 0.3.0 (2015-06-17)

- Issue #46: If VCS type is set use that rather than go get.
- Issue #45: Added git fastpath if configured ref or tag matches current
  one. (via roblillack)
- Issue #30: Added support for changed VCS type to a git repo. (thanks roblillack)
- Issue #42: Fixed update for new dependencies where repo not configured.
  (thanks roblillack)
- Issue #25: Added GOOS and GOARCH support.
- Issue #35: Updated documentation on what update from existing repos means
- Issue #37: Added support to import from GPM and Godep
- Issue #36: Added example for shell (bash/zsh) prompt to show the current
  GOPATH. (thanks eAndrius)
- Issue #31: The local Go bin should be higher precedence in the
  system's PATH (via jarod).
- Issue #28: Use HTTPS instead of HTTP for git and hg. (Thanks chendo)
- Issue #26: 'glide gopath' is smarter. It now looks for glide.yaml.
- Issue #24: Trim whitespace off of package names. (Thanks roblillack)

# Release 0.2.0 (2014-10-03)

- Issue #15, #18: `glide guess` can guess dependencies for an existing
  repo. (HUGE thanks to dz0ny)
- Issue #14: Glide fails now when YAML is invalid.
- Issue #13: cli.go added to Makefile (via roblillack)
- Issue #12: InitGlide takes YAML file now
- Issue #9: Fixed handling of $SHELL (Thanks roblillack)
- Issue #10: Symbolic link uses a relative path now (Thanks roblillack)
- Issue #5: Build step is deferred when 'go get' is used to fetch
  packages. (Thanks gsalgado)
- Issue #11: Add GOBIN to glide environment (via dz0ny)
- Typos fixed (#17 by lamielle, #16 by roblillack)
- Moved the CLI handling to cli.go (github.com/codegangsta/cli)
