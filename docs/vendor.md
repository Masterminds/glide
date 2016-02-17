# Vendor Directories

With the release of Go 1.5 the `vendor/` directory was added to the resolution locations for a dependent package in addition to the `GOPATH` and `GOROOT`. Prior to Go 1.6 you needed to opt-in before Go would look there by setting the environment variable `GO15VENDOREXPERIMENT=1`. In Go 1.6 this is an opt-out feature.

_Note, even if you use the `vendor/` directories your codebase needs to be inside the `GOPATH`. With the `go` toolchain there is no escaping the `GOPATH`._

The resolution locations for a dependent package are:

* The `vendor/` directory within the current package.
* Walk up the directory tree looking for the package in a parents `vendor/` directory.
* Look for the package in the `GOPATH`.
* Use the package in the `GOROOT` (where the standard library package reside) if present.

## Recommendations

Having worked with the `vendor/` directories since they were first released we've come to some conclusions and recommendations. Glide tries to help you with these.

1. Libraries (codebases without a `main` package) should not store outside packages in a `vendor/` folder in their VCS unless they have a specific reason and understand why they're doing it.
2. In applications (codebases with a `main` package) there should only be one `vendor/` directory at the top level of the codebase.

There are some important reasons for these recommendations.

* Each instance of a package, even the same package at the same version, in the directory structure will be in the resulting binaries. If everyone stores their own dependencies separately this will quickly lead to **binary bloat**.
* Instances of a type created from a package in one location are **not compatible** with the same package, even at the exact same version, in another location. [You can see for yourself](https://github.com/mattfarina/golang-broken-vendor). That means loggers, database connections, and other shared instances won't work.

Because of this Glide flattens the dependency tree into a single top level `vendor/` directory. If a package happens to have some dependencies in their own `vendor/` folder the `go` tool will properly resolve that version.

## Why Use A `vendor` Directory?

If we already have the `GOPATH` to store packages why is there a need for a `vendor/` directory? This is a perfectly valid question.

What if multiple applications in the `GOPATH` use different versions of the same package? This is a valid problem that's both been encountered in Go applications and widely seen in languages that have been around for a lot longer.

The `vendor/` directory allows differing codebases to have their own version available without having to be concerned with another codebase that needs a different version interfering with the version it needs. It provides a level of separation for each project.
