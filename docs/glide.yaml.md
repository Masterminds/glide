# The glide.yaml File

The `glide.yaml` file contains information about the project and the dependent packages. Here the elements of the `glide.yaml` file are outlined.

    package: github.com/Masterminds/glide
    homepage: https://masterminds.github.io/glide
    license: MIT
    owners:
    - name: Matt Butcher
      email: technosophos@gmail.com
      homepage: http://technosophos.com
    - name: Matt Farina
      email: matt@mattfarina.com
      homepage: https://www.mattfarina.com
    ignore:
    - appengine
    excludeDirs:
    - node_modules
    import:
    - package: gopkg.in/yaml.v2
    - package: github.com/Masterminds/vcs
      version: ^1.2.0
      repo:    git@github.com:Masterminds/vcs
      vcs:     git
    - package: github.com/codegangsta/cli
    - package: github.com/Masterminds/semver
      version: ^1.0.0
    testImport:
    - package: github.com/arschles/assert

These elements are:

- `package`: The top level package is the location in the `GOPATH`. This is used for things such as making sure an import isn't also importing the top level package.
- `homepage`: To find the place where you can find details about the package or applications. For example, http://k8s.io
- license: The license is either an [SPDX license](http://spdx.org/licenses/) string or the filepath to the license. This allows automation and consumers to easily identify the license.
- `owners`: The owners is a list of one or more owners for the project. This can be a person or organization and is useful for things like notifying the owners of a security issue without filing a public bug.
- `ignore`: A list of packages for Glide to ignore importing. These are package names to ignore rather than directories.
- `excludeDirs`: A list of directories in the local codebase to exclude from scanning for dependencies.
- `import`: A list of packages to import. Each package can include:
    - `package`: The name of the package to import and the only non-optional item. Package names follow the same patterns the `go` tool does. That means:
        - Package names that map to a VCS remote location end in .git, .bzr, .hg, or .svn. For example, `example.com/foo/pkg.git/subpkg`.
        - GitHub, BitBucket, Launchpad, IBM Bluemix Services, and Go on Google Source are special cases that don't need the VCS extension.
    - `version`: A semantic version, semantic version range, branch, tag, or commit id to use. For more information see the [versioning documentation](versions.md).
    - `repo`: If the package name isn't the repo location or this is a private repository it can go here. The package will be checked out from the repo and put where the package name specifies. This allows using forks.
    - `vcs`: A VCS to use such as git, hg, bzr, or svn. This is only needed when the type cannot be detected from the name. For example, a repo ending in .git or on GitHub can be detected to be Git. For a repo on Bitbucket we can contact the API to discover the type.
    - `subpackages`: A record of packages being used within a repository. This does not include all packages within a repository but rather those being used.
    - `os`: A list of operating systems used for filtering. If set it will compare the current runtime OS to the one specified and only fetch the dependency if there is a match. If not set filtering is skipped. The names are the same used in build flags and `GOOS` environment variable.
    - `arch`: A list of architectures used for filtering. If set it will compare the current runtime architecture to the one specified and only fetch the dependency if there is a match. If not set filtering is skipped. The names are the same used in build flags and `GOARCH` environment variable.
- `testImport`: A list of packages used in tests that are not already listed in `import`. Each package has the same details as those listed under import.
