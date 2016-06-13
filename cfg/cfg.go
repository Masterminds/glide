// Package cfg handles working with the Glide configuration files.
//
// The cfg package contains the ability to parse (unmarshal) and write (marshal)
// glide.yaml and glide.lock files. These files contains the details about
// projects managed by Glide.
//
// To convert yaml into a cfg.Config instance use the cfg.ConfigFromYaml function.
// The yaml, typically in a glide.yaml file, has the following structure.
//
//     package: github.com/Masterminds/glide
//     homepage: https://masterminds.github.io/glide
//     license: MIT
//     owners:
//     - name: Matt Butcher
//       email: technosophos@gmail.com
//       homepage: http://technosophos.com
//     - name: Matt Farina
//       email: matt@mattfarina.com
//       homepage: https://www.mattfarina.com
//     ignore:
//     - appengine
//     excludeDirs:
//     - node_modules
//     import:
//     - package: gopkg.in/yaml.v2
//     - package: github.com/Masterminds/vcs
//       version: ^1.2.0
//       repo:    git@github.com:Masterminds/vcs
//       vcs:     git
//     - package: github.com/codegangsta/cli
//     - package: github.com/Masterminds/semver
//       version: ^1.0.0
//
// These elements are:
//
//    - package: The top level package is the location in the GOPATH. This is used
//      for things such as making sure an import isn't also importing the top level
//      package.
//    - homepage: To find the place where you can find details about the package or
//      applications. For example, http://k8s.io
//    - license: The license is either an SPDX license string or the filepath to the
//      license. This allows automation and consumers to easily identify the license.
//    - owners: The owners is a list of one or more owners for the project. This
//      can be a person or organization and is useful for things like notifying the
//      owners of a security issue without filing a public bug.
//    - ignore: A list of packages for Glide to ignore importing. These are package
//      names to ignore rather than directories.
//    - excludeDirs: A list of directories in the local codebase to exclude from
//      scanning for dependencies.
//    - import: A list of packages to import. Each package can include:
//        - package: The name of the package to import and the only non-optional item.
//        - version: A semantic version, semantic version range, branch, tag, or
//          commit id to use.
//        - repo: If the package name isn't the repo location or this is a private
//          repository it can go here. The package will be checked out from the
//          repo and put where the package name specifies. This allows using forks.
//        - vcs: A VCS to use such as git, hg, bzr, or svn. This is only needed
//          when the type cannot be detected from the name. For example, a repo
//          ending in .git or on GitHub can be detected to be Git. For a repo on
//          Bitbucket we can contact the API to discover the type.
//    - devImport: A list of development packages. Each package has the same details
//      as those listed under import.
package cfg
