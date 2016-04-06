# Release 0.10.2 (2016-04-06)

- Issue #362: Updated docs on how -update-vendored works to help avoid confusion.
- Fixed #371: Warn when name/location mismatch.
- Fixed #290: On windows Glide was sometimes pulls in current project (thanks tzneal).
- Fixed #361: Handle relative imports (thanks tmm1).
- Fixed #373: Go 1.7 context package import issues.

# Release 0.10.1 (2016-03-25)

- Fixed #354: Fixed a situation where a dependency could be fetched when
  set to ignore.

# Release 0.10.0 (2016-03-24)

- Issue #293: Added support for importing from Gomfile's (thanks mcuelenaere).
- Issue #318: Opt-In to strip VCS metadata from vendor directory.
- Issue #297: Adds exclude property for directories in local codebase to exclude
  from scanning.
- Issue #301: Detect version control type from scp style paths (e.g. git@) and
  from scheme types (e.g., git://).
- Issue #339: Add ability to remove nested vendor and Godeps workspaces
  directories. Note, if Godeps rewriting occured it is undone. The Godeps handling
  is deprecated from day one and will be removed when most Godeps projects have
  migrated to vendor folder handling.
- Issue #350: More detailed conflict information (commit metadata displayed).
- Issue #351: Move to Gitter for chat.
- Issue #352: Make Glide installable. The dependencies are checked into the
  `vendor` folder.

# Release 0.9.3 (2016-03-09)

- Fixed #324: Glide tries to update ignored package

# Release 0.9.2 (2016-03-08)

- Fixed issue on #317: Some windows calls had the improper path separator.
- Issue #315: Track updated packages to avoid duplicated work (in part by
  thockin, thanks).
- Fixed #312: Don't double-print SetVersion() failure (thanks thockin).
- Fixed #311: Don't process deps if 'get' was a non-operation (thanks thockin).
- Issue #307: Moving 'already set' to a debug message to cleanup output
  (thanks thockin).
- Fixed #306: Don't call SetVersion twice. There was a place where it was called
  twice in a logical row (thanks thockin).
- Fixed #304: Glide tries to update ignored packages.
- Fixed #302: Force update can cause a panic.

# Release 0.9.1 (2016-02-24)

- Fixed #272: Handling appengine special package case.
- Fixed #273: Handle multiple packages in the same directory but handling
  build tags used in those packages.
- Added documentation explaining how import resolution works.
- Fixed #275 and #285: Empty directories as package locations reporting errors.
  Improved the UX and handle the errors.
- Fixed #279: Added Go 1.7 support that no longer has GO15VENDOREXPERIMENT.
- Issue #267: Added `os` and `arch` import properties to the documentation.
- Fixed #267: Glide was only walking the import tree based on build flags for
  the current OS and Arch. This is a problem for systems like docker that have
  variation built in.

# Release 0.9.0 (2016-02-17)

- Fixed #262: Using correct query string merging for go-get queries (thanks gdm85).
- Fixed #251: Fixed warning message (thanks james-lawrence).
- Adding support for IBM JazzHub.
- Fixes #250: When unable to retrieve or set version on a dependency now erroring
  and exiting with non-0 exit code.
- Issue #218: Added `glide rm` command.
- Fixed #215: Under some error conditions the package resolver could get into
  an infinite loop.
- Issue #234: Adding more options to the glide.yaml file including license,
  owners, homepage, etc. See the docs for more detail.
- Issue #237: Added Read The Docs support and initial docs. http://glide.readthedocs.org
- Issue #248: Uses go env to get value of GO15VENDOREXPERIMENT due to 1.6 enabling
  by default.
- Issue #240: Glide only scans used imports rather than all paths in the tree.
  The previous behavior is available via a flag.
- Fixed #235: Glide on windows writing incorrect slashes to files.
- Fixed #227: Fixed ensure when multiple gopaths.
- Refactored Glide
  - Many features broken out into packages. All but `action/` can be
    used as libraries.
  - Cookoo is not used anymore
  - The `action/` package replaces `cmd/`

# Release 0.8.3 (2015-12-30)

- Issue #198: Instead of stopping `glide install` for a hash failures providing
  a warning. Failed hash check is currently too aggressive.
- Fixed #199: `glide up` on Windows unable to detect dependencies when GOPATH
  and GOROOT on a different drive or when GOROOT ends in a path separator.
- Fixed #194: `glide up` stalling on Windows due to POSIX path separators and
  path list separators being used.
- Fixed #185 and #187: Inaccurate hash being generated for lock file with nested
  version ranges.
- Fixed #182 and #183: Caching on go-import lookups mishandled some prefixes.
- Fixed issue in deduping and sub-package names.
- Fixed #189: nested dependencies that do not contain VCS information were not
  being updated properly when --updated-vendored was being used.
- Fixed #186: glide up PACKAGE was failing to generate a proper glide.lock file.

# Release 0.8.2 (2015-12-21)

- Fixed #169: cookoo git url has auth info. Makes glide unbuildable for
  environments not setup for GitHub.
- Fixed #180: the hash in the glide.lock file was not being properly calculated.
- Fixed #174: glide get was causing an error when the flag --updated-vendored
  was being used.
- Fixed #175: glide get when the GOPATH isn't setup properly could end up in
  an infinite loop.

# Release 0.8.1 (2015-12-15)

- Fixed #163: Was detecting std lib packages when the GOROOT was different at
  runtime than compile time.
- Fixed #165: glide update panics with --no-recursive option.
- Added back zip build option to build scripts. This is useful for some
  environments.

# Release 0.8.0 (2015-12-10)

- Issues #156 and #85: Added lockfile support (glide.lock). This file records
  commit id pinned versions of the entire dependency tree. The `glide install`
  command installs the pinned dependencies from the `glide.lock` file while
  `glide update` updates the tree and lockfile. Most people should use `glide
  install` unless they want to intentionally updated the pinned dependencies.
  `glide install` is able to use concurrency to more quickly install update.
- Issues #33 and #159: Glide notifies if a dependency checkout has uncomitted
  changes.
- Issue #146: Glide scans projects not managed by a dependency manager, fetches
  their dependencies, and pins them in the glide.lock file.
- Issue #99: Glide `get` pins dependencies by default and allows a version to
  be passed in. For example, `glide get github.com/Masterminds/convert#^1.0.0`
  will fetch `github.com/Masterminds/convert` with a version of `^1.0.0`.
- Issue #155: Copying packages from the `GOPATH` is now opt-in.

# Release 0.7.2 (2015-11-16)

- Fixed #139: glide.yaml file imports being reordered when file written.
- Fixed #140: packages in glide.yaml were no longer being deduped.

# Release 0.7.1 (2015-11-10)

- Fixed #136: Fixed infinite recursion in list and tree commands.
- Fixed issue where glide guess listed a null parent.
- Fixed #135: Hard failure when home directory not found for cache.
- Fixed #137: Some messages not ending in "\n".
- Fixed #132 and #133: Build from source directions incorrect (thanks hyPiRion).

# Release 0.7.0 (2015-11-02)

- Fixed #110: Distribution as .tag.gz instead of .zip.
- Issue #126: Added --no-color option to remove color for systems that do not
  work well with color codes (thanks albrow).
- Added caching functionality (some opt-in).
- Added global debug flag.
- Moved yaml parsing and writing to gopkg.in/yaml.v2 and separated
  config handling into separate package.
- Better godep import handling.
- Fixed #98: Godep command name fix (thanks jonboulle).
- #52 and #114: Add semantic version (SemVer) support.
- #108: Flatten the dependency tree by default.
- Fixed #107: Allow `glide get` to retrieve insecure packages with `--insecure`
  flag.
- #105: Import commands accept a filename with the `-f` flag.
- Fixed #97: Fixed misspellings (thanks jonboulle).
- #96: Allow multiple packages in `glide get`.
- #92: Added support to `glide update` to only update a specific package.
- #91: `glide list` now displays if a pkg is in vendor, GOPATH, or missing.
- Issue #89: More robust GOPATH handling (thanks gcmt).
- Fixed #65: Hg commands were not checking out the codebase on the first update.
- Fixed #95: Added more detail for errors previously reporting "Oops! exit
  status 128".
- Fixed #86 and #71: Imported package names including a sub-package were checked
  out to the wrong location. They are not checked out to the right place and
  multiple instances of the top level repo are merged with error checking.

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
