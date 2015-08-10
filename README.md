# Glide: Managing Go Workspaces With Ease

*Manage your vendor and vendored packages with ease.* Glide is a tool for
managing the `vendor` directory within a Go package. This feature, first
introduced in Go 1.5, allows each package to have a `vendor` directory
containing dependent packages for the project. These vendor packages can be
installed by a tool (e.g. glide), similar to `go get` or they can be vendored and
distributed with the package.

[![Build Status](https://travis-ci.org/Masterminds/glide.svg)](https://travis-ci.org/Masterminds/glide)

### Features

* Ease dependency management
* Support **versioning packages**
* Support **aliasing packages** (e.g. for working with github forks)
* Remove the need for munging import statements
* Work with all of the `go` tools
* Support the VCS tools that Go supports:
    - git
    - bzr
    - hg
    - svn
* Support custom local and global plugins (see docs/plugins.md)

## How It Works

The dependencies for a project are listed in a `glide.yaml` file. This can
include a version, VCS, repository location (that can be different from the
package name), etc. When `glide up` is run it downloads the packages (or updates)
to the `vendor` directory. It then recursively walks through the downloaded
packages looking for those with a `glide.yaml` file that don't already have
a `vendor` directory and installing their dependencies to their `vendor`
directories.

 A projects is structured like this:

```
- myProject (Your project)
  |
  |-- glide.yaml
  |
  |-- main.go (Your main go code can live here)
  |
  |-- mySubpackage (You can create your own subpackages, too)
  |    |
  |    |-- foo.go
  |
  |-- vendor
       |-- github.com
            |
            |-- Masterminds
                  |
                  |-- ... etc.
```

*Take a look at [the Glide source code](http://github.com/Masterminds/glide)
to see this philosophy in action.*

## Install

On Mac OS X you can install the latest release via [Homebrew](https://github.com/Homebrew/homebrew):

```
$ brew install glide
```

[Binary packages](https://github.com/Masterminds/glide/releases) are available for Mac and Linux.

To build from source you can:

1. Clone this repository and change directory into it
2. Run `make bootstrap`

This will leave you with `./glide`, which you can put in your `$PATH` if
you'd like. (You can also take a look at `make install` to install for
you.)

The Glide repo has now been configured to use glide to
manage itself, too.

## Usage

```
$ glide create                            # Start a new workspaces
$ open glide.yaml                         # and edit away!
$ glide get github.com/Masterminds/cookoo # Get a package and add to glide.yaml
$ glide install                           # Install packages and dependencies
# work, work, work
$ go build                                # Go tools work normally
$ glide up                                # Update to newest versions of the package
```

Check out the `glide.yaml` in this directory, or examples in the `docs/`
directory.

### glide create

Initialize a new workspace. Among other things, this creates a stub
`glide.yaml`

```
$ glide create
[INFO] Initialized. You can now edit 'glide.yaml'
```

### glide get [package name]

You can download package to your `vendor` directory and have it added to your
`glide.yaml` file with `glide get`.

```
$ glide get github.com/Masterminds/cookoo
```

### glide up (aliased to update and install)

Download or update all of the libraries listed in the `glide.yaml` file and put
them in the `vendor` directory. It will also recursively walk through the
dependency packages doing the same thing if no `vendor` directory exists.

```
$ glide up
```

### glide rebuild

Re-run `go install` on the packages in the `glide.yaml` file. This
(along with `glide install` and `glide update`) pays special attention
to the contents of the `subpackages:` directive in the YAML file.

```
$ glide rebuild
[INFO] Building dependencies.
[INFO] Running go build github.com/kylelemons/go-gypsy/yaml
[INFO] Running go build github.com/Masterminds/cookoo/cli
[INFO] Running go build github.com/Masterminds/cookoo
```

### glide help

Print the glide help.

```
$ glide help
```

### glide version

Print the version and exit.

```
$ glide version
0.0.2-3-g4ac84b4
```

### glide.yaml

The `glide.yaml` file does two critical things:

1. It names the current package
2. It declares external dependencies

A brief `glide.yaml` file looks like this:

```yaml
package: github.com/Masterminds/glide
import:
  - package: github.com/kylelemons/go-gypsy
  - package: github.com/Masterminds/cookoo
    vcs: git
    ref: master
    repo: git@github.com:Masterminds/cookoo.git
```

The above tells `glide` that...

1. This package is named `github.com/Masterminds/glide`
2. That this package depends on two libraries.


The first library exemplifies a minimal package import. It merely gives
the fully qualified import path.

When Glide reads the definition for the second library, it will get the repo
from the source in `repo`, checkout the master branch, and put it in
`github.com/Masterminds/cookoo` in the `vendor` directory. (Note that `package`
and `repo` can be completely different)

**TIP:** The ref is VCS dependent and can be anything that can be checked out.
For example, with Git this can be a branch, tag, or hash. This varies and depends
on what's supported in the VCS.

**TIP:** In general, you are advised to use the *base package name* for
importing a package, not a subpackage name. For example, use
`github.com/kylelemons/go-gypsy` and not
`github.com/kylelemons/go-gypsy/yaml`.

### Controlling package and subpackage builds

In addition to fetching packages, Glide builds the packages with `go
install`. The YAML file can give special instructions about how to build
a package. Example:

```yaml
package: github.com/technosophos/glide
import:
  - package: github.com/kylelemons/go-gypsy
    subpackage: yaml
  - package: github.com/Masterminds/cookoo
    subpackage:
      - .
      - cli
      - web
  - package: github.com/crowdmob/amz
    subpackage: ...
```

According to the above, the following packages will be built:

1. The `go-gypsy/yaml` package
2. The `cookoo` package (`.`), along with `cookoo/web` and `cookoo/cli`
3. Everything in `awz` (`...`)

See the `docs/` folder for more examples.

## Supported Version Control Systems

The Git, SVN, Mercurial (Hg), and Bzr source control systems are supported. This
happens through the [vcs package](https://github.com/masterminds/go-vcs).

## Troubleshooting

**Q: bzr (or hg) is not working the way I expected. Why?**

These are works in progress, and may need some additional tuning. Please
take a look at the [vcs package](https://github.com/masterminds/go-vcs). If you
see a better way to handle it please let us know.

**Q: Should I check `vendor/` into version control?**

That's up to you. It's not necessary, but it may also cause you extra
work and lots of extra space in your VCS.

**Q: How do I import settings from GPM or Godep?**

Glide can import from GPM's `Godeps` file format or from Godep's
`Godeps/Godeps.json` file format.

For GPM, use `glide import gpm`.

For Godep, use `glide import godep`.

Each of these will merge your existing `glide.yaml` file with the
dependencies it finds for those managers, and then emit the file as
output. **It will not overwrite your glide.yaml file.**

You can write it to file like this:

```
$ glide import godep > new-glide.yaml
```

**Q: Can Glide fetch a package based on OS or Arch?**

A: Yes. Using the `os` and `arch` fields on a `package`, you can specify
which OSes and architectures the package should be fetched for. For
example, the following package will only be fetched for 64-bit
Darwin/OSX systems:

```yaml
- package: some/package
  os:
    - darwin
  arch:
    - amd64
```

The package will not be fetched for other architectures or OSes.

## LICENSE

This package is made available under an MIT-style license. See
LICENSE.txt.

## Thanks!

We owe a huge debt of gratitude to the [GPM and
GVP](https://github.com/pote/gpm) projects, which
inspired many of the features of this package. If `glide` isn't the
right Go project manager for you, check out those.

The Composer (PHP), npm (JavaScript), and Bundler (Ruby) projects all
inspired various aspects of this tool, as well.

## The Name

Aside from being catchy, "glide" is a contraction of "Go Elide". The
idea is to compress the tasks that normally take us lots of time into a
just a few seconds.
