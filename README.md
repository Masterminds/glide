# Glide: Managing Go Workspaces With Ease

*Never vendor again.* Glide is a tool for managing Go package dependencies and
[Go workspaces](http://golang.org/doc/code.html#GOPATH). Subscribing to the
view that each project should have its
own GOPATH, Glide provides tools for versioning Go libraries and
managing the environment in which your normal Go tools run.

[![Build Status](https://travis-ci.org/Masterminds/glide.svg)](https://travis-ci.org/Masterminds/glide)

### Features

* Manage project-specific `GOPATH`s
* Ease dependency management
* Support **versioning packages**
* Support **aliasing packages** (e.g. for working with github forks)
* Remove the need for "vendoring" or munging import statements
* Work with all of the `go` tools
* Support the VCS tools that Go supports:
    - git
    - bzr
    - hg
    - svn
* Support custom local and global plugins (see docs/plugins.md)

## How It Works

Glide is an opinionated tool for managing Go workspaces. Glide associates
a GOPATH to a particular workspace with its own particular dependencies.
And it assumes that each project has its main source code and also some
number of dependent packages.

Projects are structured like this:

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
  |-- _vendor (This is $GOPATH)
       |
       |-- bin
       |
       |-- src
            |
            |-- github.com
                  |
                  |-- Masterminds
                       |
                       |-- ... etc.
```

Through some trickery, the GOPATH is set to `_vendor`, but the go tools
will still find `main.go` and subpackages. Make sure, though, that you
set the name of your package in `glide.yaml`.

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
$ glide create    # Start a new workspaces
$ glide in        # Switch into the workspace
$ open glide.yaml # and edit away!
$ glide install   # Install packages and dependencies
# work, work, work
$ go build        # Go tools work normally
$ glide update    # Update to newest versions of the package
$ exit            # Exit the glide session (started with glide in)
```

Check out the `glide.yaml` in this directory, or examples in the `docs/`
directory.

### glide create

Initialize a new workspace. Among other things, this creates a stub
`glide.yaml`

```
$ glide create
[INFO] Your new GOPATH is /Users/mbutcher/Code/glide/docs/_vendor. Run 'glide gopath' to see it again.
[INFO] Initialized. You can now edit 'glide.yaml'
```

**If you set your GOPATH in your shell's profile or RC scripts, you may
need to tweak those settings. See the Troubleshooting section below.**

### glide in

Configure an interactive shell for working in a project. This configures
the GOPATH and so on.

```
$ glide in
>> You are now gliding into a new shell. To exit, type 'exit'
$ echo $GOPATH
/Users/mbutcher/Code/glide/_vendor
$ exit
>> Exited glide shell
$
```

For ease of use, there's a special variant of
`glide in` called `glide into`:

```
glide into /foo/bar
```

The above will change directories into `/foo/bar`, make sure it's a Go
workspace, and then launch a new Glide shell.

**If you set your GOPATH in your shell's profile or RC scripts, you may
need to tweak those settings. See the Troubleshooting section below.**

### glide install

Download all of the libraries listed in the `glide.yaml` file and put
them where they should go.

```
$ glide install
```

### glide update

Update all of the existing repositories. If a new new repository has
been added to the YAML file, try to download that, too.

```
$ glide update
[INFO] Updating github.com/kylelemons/go-gypsy/yaml with 'go get -u'
[INFO] Updating github.com/Masterminds/cookoo with Git (From git@github.com:Masterminds/cookoo.git)
Fetching origin
[INFO] Updating github.com/aokoli/goutils with 'go get -u'
[INFO] Updating github.com/crowdmob/goamz with Git (From git@github.com:technosophos/goamz.git)
Fetching origin
[INFO] Set version to github.com/Masterminds/cookoo to master
[INFO] Looks like /Users/mbutcher/Code/glide/_vendor/src/github.com/aokoli/goutils is a Git repo.
[INFO] Set version to github.com/aokoli/goutils to the latest
[INFO] Set version to github.com/crowdmob/goamz to the latest
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

### glide gopath

Emit the GOPATH to this project. Useful for things like `GOPATH=$(glide
gopath)`.

```
$ glide gopath
/Users/mbutcher/Code/glide/_vendor
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
package: github.com/technosophos/glide
import:
  - package: github.com/kylelemons/go-gypsy
  - package: github.com/Masterminds/cookoo
    vcs: git
    ref: master
    repo: git@github.com:Masterminds/cookoo.git
```

The above tells `glide` that...

1. This package is named `github.com/technosophos/glide`
2. That this package depends on two libraries.


The first library exemplifies a minimal package import. It merely gives
the fully qualified import path. Glide will use `go get` to initially
fetch it.

The second library forgoes `go get` and uses `git` directly. When Glide
reads this definition, it will get the repo from the source in `repo`
and then checkout the master branch, and put it in
`github.com/Masterminds/cookoo` in the GOPATH. (Note that `package` and
`repo` can be completely different)

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

### Displaying glide environment indicator in bash

To display an indicator in a custom Bash prompt whenever you are
"in" the glide environment, add the following check to your $PS1
environment variable (likely in `.bashrc`):

```bash
$(if [ "$ALREADY_GLIDING" = "1" ]; then echo " (gliding)"; fi)
```

Example:

```bash
export PS1='\u@\h \w$(if [ "$ALREADY_GLIDING" = "1" ]; then echo " (gliding)"; fi) \n$ '
```

Result:

```
user@hostname ~/your/go/project
$ glide in
>> You are now gliding into a new shell. To exit, type 'exit'
user@hostname ~/your/go/project (gliding)
$
```

## Supported Version Control Systems

Anything supported by `go get` works out of the box. By default, we use
'go get' to fetch and install dependencies. However, if you use
`repository` or `ref` statements in your `glide.yaml` file, the native
client will be used directly.

Support for these is a little harder, and requires some expertise in
each system.

### Fully supported:

- git

### Supported, but not tested well:

- bzr: All operations supported, but maybe not ideally.
- hg: All operations supported, but maybe not ideally.
- svn: Checkout and update are supported. Checkout by branch or tag is
  done by setting the `repository` URL appropriately. Checkout by `ref`
  supports revision numbers and symbolic references.

See [docs/vcs.md](docs/vcs.md) for more info.

## Troubleshooting

**Q: When I `glide in` a project, my $GOPATH goes to the default.
Why?**

If you're shell's startup (`.profile`, `.bashrc`, `.zshrc`) sets a
default `$GOPATH`, this will override the `GOPATH` that glide sets. The
simple work-around is to use this in your profile:

```bash
if [ "" = "${GOPATH}" ]; then
  export GOPATH="/some/dir"
fi
```

This will only set a GOPATH if one does not exist. Alternately, if you want to
set the GOPATH when you're not using `glide in` or `glide into` try the following:

```bash
if [ "" = "${ALREADY_GLIDING}" ]; then
  export GOPATH="/some/dir"
fi
```

**Q: bzr (or hg) is not working the way I expected. Why?**

These are works in progress, and may need some additional tuning. Please
take a look at `cmd/bzr.go` and `cmd/hg.go` to see what we do. If you
can make it better, please submit a patch.

**Q: When I 'glide in', I want to do something cooler than what you do.
How?**

You can use `incmd: some custom command` in your glide.yaml file.
Example:

```
incmd: bash -l
```

With the above, running `glide in` will start a new Bash shell
simulating a login environment.

**Q: I don't want to use 'glide in'. How do I set my GOPATH?**

You may explicitly set the GOPATH like this:

```bash
export GOPATH=$(glide gopath)
```

The command `glide gopath` will emit the correct path to set as GOPATH.

**Q: Is using the Glide GOPATH required? Do I have to use `_vendor`?**

No, it is not required, and you do not need to use `_vendor`. You may
choose to use another GOPATH manager, like
[GVP](http://github.com/pote/gvp), or you may simply manage GOPATH on
your own.

**Q: Should I check `_vendor` into version control?**

That's up to you. It's not necessary, but it may also cause you extra
work and lots of extra space in your VCS.

**Q: How can I get my `_vendor` path to work with Sublime Text and GoSublime?**

GoSublime uses an application wide GOPATH. If you want a different GOPATH codebase
set them up as different projects. Then, in the project settings (your `.sublime-project`
file) add an entry to set the GOPATH. For example:

```json
{
    "settings": {
        "GoSublime": {
            "env": {
                "GOPATH": "$HOME/path/to/project/_vendor"
            }
        }
    },
    "folders":
    [
        {
            "follow_symlinks": true,
            "path": "."
        }
    ]
}
```
Once you've done this feature like autocomplete will work.

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
