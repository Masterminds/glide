# Version Control Systems

This document explains how we'd like Glide's VCS integration to work.

## Call for Help!

We're pretty comfortable with `go get` and `git`, but we're a little out
of our element with `hg` and `bzr`. We're not sure that our
implementation is the best. If you have experience with these, please
take a look at `cmd/hg.go` and `cmd/bzr.go`. In particular, see *Goal 3*
below.

## Goal 1: Use 'go get' when it makes sense

There's no need to re-invent the wheel, so we use `go get` when it makes
sense to do so.

The most obvious case for using `go get` is when the only desired Glide
action for a package is mere installation:

```yaml
import:
  - package: github.com/technosophos/foo
```

In this case, Glide uses `go get` to install, `go get -u` to update, and
nothing else.

In most cases, it also makes sense to use `go get` when the only
additional behavior is setting a reference (`ref`, version):

```yaml
import:
  - package: github.com/technosophos/foo
    ref: 1.1.1
```

In this case, we use `go get` to install, `go get -u` to update, and
then guess the repo type. Since this is a `git` repo, `git checkout`
will be used to get the particular tag.

## Goal 2: Support popular VCS systems directly

We would like to support the following VCS systems *fully*:

- git
- bzr
- hg
- svn

What do we mean by 'fully'? We mean:

- Be able to create (clone) a local copy of the repository
- Be able to update the local copy to the latest commit(s)
- Be able to check out...
  * branches
  * tags
  * individual commits

## Goal 3: 'ref' and 'repo' can identify a branch, tag, or commit

To keep things simple, we'd ideally like to use only two YAML
configuration directives to identify the precise version to install:

- repository (`repo`)
- reference (`ref`)

We don't feel like we need to create an exactly identical process for
each VCS. For example, a `git` reference can refer to a commit
(`a34d523`), a tag (`1.1.1`) or branch (`develop`).

Subversion, on the other hand, uses URLs (`repo`) to indicate tags
and branches, and references (`ref`) for commit numbers (`321`).
