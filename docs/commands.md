# Commands

The following are the Glide commands, most of which are to help yoy manage your workspace.

## glide create (aliased to init)

Initializes a new workspace. Among other things, this creates a `glide.yaml` file while attempting to guess the packages and versions to put in it. For example, if your project is using Godep it will use the versions specified there. Glide is smart enough to scan your codebase and detect the imports being used whether they are specified with another package manager or not.

    $ glide init
    [INFO] Generating a YAML configuration file and guessing the dependencies
    [INFO] Attempting to import from other package managers (use --skip-import to skip)
    [INFO] Found reference to github.com/BurntSushi/toml
    [INFO] Found reference to github.com/Masterminds/semver
    [INFO] Found reference to github.com/Masterminds/sprig
    [INFO] Found reference to github.com/Masterminds/vcs
    [INFO] Found reference to github.com/aokoli/goutils
    [INFO] Found reference to github.com/codegangsta/cli
    [INFO] Found reference to github.com/deis/pkg/prettyprint
    [INFO] Found reference to github.com/ghodss/yaml
    [INFO] Found reference to github.com/google/go-github/github
    [INFO] Found reference to github.com/pborman/uuid
    [INFO] Found reference to golang.org/x/crypto/nacl/box
    [INFO] Adding sub-package ssh/terminal to golang.org/x/crypto
    [INFO] Found reference to gopkg.in/yaml.v2
    ...

## glide get [package name]

You can download one or more packages to your `vendor` directory and have it added to your
`glide.yaml` file with `glide get`.

    $ glide get github.com/Masterminds/cookoo

When `glide get` is used it will introspect the listed package to resolve its dependencies including using Godep, GPM, Gom, and GB config files.

The `glide get` command can have a [version or range](versions.md) passed in with the package name. For example,

    $ glide get github.com/Masterminds/cookoo#^1.2.3

The version is separated from the package name by an anchor (`#`).

## glide update (aliased to up)

Download or update all of the libraries listed in the `glide.yaml` file and put them in the `vendor` directory. It will also recursively walk through the dependency packages doing the same thing if no `vendor` directory exists.

    $ glide up

This will recurse over the packages looking for other projects managed by Glide, Godep, GB, Gom, and GPM. When one is found those packages will be installed as needed.

A `glide.lock` file will be created or updated with the dependencies pinned to specific versions. For example, if in the `glide.yaml` file a version was specified as a range (e.g., `^1.2.3`) it will be set to a specific commit id in the `glide.lock` file. That allows for reproducible installs (see `glide install`).

If you want to use `glide up` to help you managed dependencies that are checked into your version control consider the flags:

* `--update-vendored` (aliased to `-u`) to update the vendored dependencies. If Glide detects a vendored dependency it will update it and leave it in a vendored state. Note, any tertiary dependencies will not be automatically vendored with this flag.
* `--strip-vcs` (aliased to `-s`) to strip VCS metadata (e.g., `.git` directories) from the `vendor` folder.
* `--strip-vendor` (aliased to `-v`) to strip nested `vendor/` directories.

For example, you can use the command:

    $ glide up -u -s

This will tell Glide to update the vendored packages and remove any VCS directories from transitive dependencies that were picked up as well.

## glide install

When you want to install the specific versions from the `glide.lock` file use `glide install`.

    $ glide install

This will read the `glide.lock` file, warning you if it's not tied to the `glide.yaml` file, and install the commit id specific versions there.

When the `glide.lock` file doesn't tie to the `glide.yaml` file, such as there being a change, it will provide an warning. Running `glide up` will recreate the `glide.lock` file when updating the dependency tree.

If no `glide.lock` file is present `glide install` will perform an `update` and generates a lock file.

## glide novendor (aliased to nv)

When you run commands like `go test ./...` it will iterate over all the subdirectories including the `vendor` directory. When you are testing your application you may want to test your application files without running all the tests of your dependencies and their dependencies. This is where the `novendor` command comes in. It lists all of the directories except `vendor`.

    $ go test $(glide novendor)

This will run `go test` over all directories of your project except the `vendor` directory.

## glide name

When you're scripting with Glide there are occasions where you need to know the name of the package you're working on. `glide name` returns the name of the package listed in the `glide.yaml` file.

## glide list

Glide's `list` command shows an alphabetized list of all the packages that a project imports.

```
$ glide list
INSTALLED packages:
	vendor/github.com/Masterminds/cookoo
	vendor/github.com/Masterminds/cookoo/fmt
	vendor/github.com/Masterminds/cookoo/io
	vendor/github.com/Masterminds/cookoo/web
	vendor/github.com/Masterminds/semver
	vendor/github.com/Masterminds/vcs
	vendor/github.com/codegangsta/cli
	vendor/gopkg.in/yaml.v2
```

## glide help

Print the glide help.

```
$ glide help
```

## glide --version

Print the version and exit.

```
$ glide --version
glide version 0.9.0
```
