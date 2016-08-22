# Commands

The following are the Glide commands, most of which are to help yoy manage your workspace.

## glide create (aliased to init)

Initialize a new workspace. Among other things, this creates a `glide.yaml` file
while attempting to guess the packages and versions to put in it. For example,
if your project is using Godep it will use the versions specified there. Glide
is smart enough to scan your codebase and detect the imports being used whether
they are specified with another package manager or not.

    $ glide create
    [INFO]	Generating a YAML configuration file and guessing the dependencies
    [INFO]	Attempting to import from other package managers (use --skip-import to skip)
    [INFO]	Scanning code to look for dependencies
    [INFO]	--> Found reference to github.com/Masterminds/semver
    [INFO]	--> Found reference to github.com/Masterminds/vcs
    [INFO]	--> Found reference to github.com/codegangsta/cli
    [INFO]	--> Found reference to gopkg.in/yaml.v2
    [INFO]	Writing configuration file (glide.yaml)
    [INFO]	Would you like Glide to help you find ways to improve your glide.yaml configuration?
    [INFO]	If you want to revisit this step you can use the config-wizard command at any time.
    [INFO]	Yes (Y) or No (N)?
    n
    [INFO]	You can now edit the glide.yaml file. Consider:
    [INFO]	--> Using versions and ranges. See https://glide.sh/docs/versions/
    [INFO]	--> Adding additional metadata. See https://glide.sh/docs/glide.yaml/
    [INFO]	--> Running the config-wizard command to improve the versions in your configuration

The `config-wizard`, noted here, can be run here or manually run at a later time.
This wizard helps you figure out versions and ranges you can use for your
dependencies.

### glide config-wizard

This runs a wizard that scans your dependencies and retrieves information on them
to offer up suggestions that you can interactively choose. For example, it can
discover if a dependency uses semantic versions and help you choose the version
ranges to use.

## glide get [package name]

You can download one or more packages to your `vendor` directory and have it added to your
`glide.yaml` file with `glide get`.

    $ glide get github.com/Masterminds/cookoo

When `glide get` is used it will introspect the listed package to resolve its dependencies including using Godep, GPM, Gom, and GB config files.

The `glide get` command can have a [version or range](versions.md) passed in with the package name. For example,

    $ glide get github.com/Masterminds/cookoo#^1.2.3

The version is separated from the package name by an anchor (`#`). If no version or range is specified and the dependency uses Semantic Versions Glide will prompt you to ask if you want to use them.

## glide update (aliased to up)

Download or update all of the libraries listed in the `glide.yaml` file and put
them in the `vendor` directory. It will also recursively walk through the
dependency packages to fetch anything that's needed and read in any configuration.

    $ glide up

This will recurse over the packages looking for other projects managed by Glide,
Godep, gb, gom, and GPM. When one is found those packages will be installed as needed.

A `glide.lock` file will be created or updated with the dependencies pinned to
specific versions. For example, if in the `glide.yaml` file a version was
specified as a range (e.g., `^1.2.3`) it will be set to a specific commit id in
the `glide.lock` file. That allows for reproducible installs (see `glide install`).

To remove any nested `vendor/` directories from fetched packages see the `-v` flag.

## glide install

When you want to install the specific versions from the `glide.lock` file use `glide install`.

    $ glide install

This will read the `glide.lock` file, warning you if it's not tied to the `glide.yaml` file, and install the commit id specific versions there.

When the `glide.lock` file doesn't tie to the `glide.yaml` file, such as there being a change, it will provide an warning. Running `glide up` will recreate the `glide.lock` file when updating the dependency tree.

If no `glide.lock` file is present `glide install` will perform an `update` and generates a lock file.

To remove any nested `vendor/` directories from fetched packages see the `-v` flag.

## glide novendor (aliased to nv)

When you run commands like `go test ./...` it will iterate over all the subdirectories including the `vendor` directory. When you are testing your application you may want to test your application files without running all the tests of your dependencies and their dependencies. This is where the `novendor` command comes in. It lists all of the directories except `vendor`.

    $ go test $(glide novendor)

This will run `go test` over all directories of your project except the `vendor` directory.

## glide name

When you're scripting with Glide there are occasions where you need to know the name of the package you're working on. `glide name` returns the name of the package listed in the `glide.yaml` file.

## glide list

Glide's `list` command shows an alphabetized list of all the packages that a project imports.

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

## glide help

Print the glide help.

    $ glide help

## glide --version

Print the version and exit.

    $ glide --version
    glide version 0.12.0

## glide mirror

Mirrors provide the ability to replace a repo location with
another location that's a mirror of the original. This is useful when you want
to have a cache for your continuous integration (CI) system or if you want to
work on a dependency in a local location.

The mirrors are stored in an `mirrors.yaml` file in your `GLIDE_HOME`.

The three commands to manager mirrors are `list`, `set`, and `remove`.

Use `set` in the form:

    glide mirror set [original] [replacement]

or

    glide mirror set [original] [replacement] --vcs [type]

for example,

   glide mirror set https://github.com/example/foo https://git.example.com/example/foo.git

   glide mirror set https://github.com/example/foo file:///path/to/local/repo --vcs git

Use `remove` in the form:

   glide mirror remove [original]

for example,

   glide mirror remove https://github.com/example/foo
