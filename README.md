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

## LICENSE

This package is made available under an MIT-style license.

## Thanks!

We owe a huge debt of gratitude to the GPM and GVP projects, which
inspired many of the features of this package. If `glide` isn't the
right Go project manager for you, check out those.

The Composer (PHP), npm (JavaScript), and Bundler (Ruby) projects all
inspired various aspects of this tool, as well.

## The Name

Aside from being catchy, "glide" is a contraction of "Go Elide". The
idea is to compress the tasks that normally take us lots of time into a
just a few seconds.

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
