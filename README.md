# Glide: Managing Go Projects With Ease

**WARNING:** Glide is a toy project right now, and should not really be
used for anything at all.

Glide is a tool for managing Go projects. It is intended to do the
following:

* Manage project-specific `GOPATH`s
* Ease dependency management
* Support versioning in packages
* Support aliasing packages (e.g. for working with github forks)
* Remove the need for "vendoring" or munging import statements
* Work with all of the `go` tools
* Support the VCS tools that Go supports:
    - git
    - bzr
    - hg
    - svn

## How It Works

Glide is an opinionated tool for managing Go projects. Glide associates
a GOPATH with a particular project with its own particular dependencies.
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

## Usage

```
$ glide init      # Start a new project
$ glide in        # Switch into the new project
$ open glide.yaml # and edit away!
$ glide install   # Install packages and dependencies
# work, work, work
$ go build        # Go tools work normally
$ glide update    # Update to newest versions of the package
$ exit            # Exit the glide session (started with glide in)
```

Check out the `glide.yaml` in this directory, or examples in the `docs/`
directory.

### glide init

Initialize a new project. Among other things, this creates a stub
`glide.yaml`

### glide in

Configure an interactive shell for working in a project. This configures
the GOPATH and so on. For ease of use, there's a special variant of
`glide in` called `glide into`:

```
glide into /foo/bar
```

The above will change directories into `/foo/bar`, make sure it's a Go
project, and then launch a new Glide shell.

### glide install

Download all of the libraries listed in the `glide.yaml` file and put
them where they should go.

### glide update

Update all of the existing repositories. If a new new repository has
been added to the YAML file, try to download that, too.

### glide gopath

Emit the GOPATH to this project. Useful for things like `GOPATH=$(glide
gopath)`.

### glide help

Print the glide help.

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

## Troubleshooting

**Q: When I `glide in` a project, my $GOPATH goes to the default.
Why?**

If you're shell's startup (`.profile`, `.bashrc`, `.zshrc`) sets a
default `$GOPATH`, this will override the `GOPATH` that glide sets. The
simple work-around is to use this in your profile:

```bash
if [ "" = "${ALREADY_GLIDING}" ]; then
  export GOPATH="/some/dir"
fi
```

The above will *only* set GOPATH if you're not using `glide in` or
`glide into`.

**Q: bzr (or hg) is not working the way I expected. Why?**

These are works in progress, and may need some additional tuning. Please
take a look at `cmd/bzr.go` and `cmd/hg.go` to see what we do. If you
can make it better, please submit a patch.

## LICENSE

This package is made available under an MIT-style license.

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

