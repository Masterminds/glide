# Glide: Managing Go Projects With Ease

**WARNING:** Glide is a toy project right now, and should not really be
used for anything at all.

Glide is a tool for managing Go projects. It is intended to do the
following:

* Manage project-specific `GOPATH`s
* Ease dependency management
* Support versioning in packages
* Remove the need for "vendoring" or munging import statements
* Support "prebuilding" of dependencies
in
And it does all of this with a simple tool and a simple JSON format.

## Usage

```
$ `glide in` # backticks are currently necessary
$ glide init
$ open glide.yaml # and edit away!
$ glide install
# work, work, work
$ glide update
$ `glide out`
```

Check out the `glide.yaml` in this directory, or examples in the `docs/`
directory.

## LICENSE

This package is made available under an MIT-style license.

## The Name

Aside from being catchy, "glide" is a contraction of "Go Elide". The
idea is to compress the tasks that normally take us lots of time into a
just a few seconds.
