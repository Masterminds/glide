# Versions and Ranges

Glide supports [Semantic Versions](http://semver.org), SemVer ranges, branches, tags, and commit ids as versions.

## Basic Ranges
A simple range is in the form `> 1.2.3`. This tells Glide to use the latest versions that's after `1.2.3`. Glide has support for the following operators:

* `=`: equal (aliased to no operator)
* `!=`: not equal
* `>`: greater than
* `<`: less than
* `>=`: greater than or equal to
* `<=`: less than or equal to

These can be combined. A `,` is an and operator and a `||` is an or operator. The or operators cause groups of and operators to be checked. For example, `">= 1.2, < 3.0.0 || >= 4.2.3"`.

## Hyphen Ranges

There are multiple shortcuts to handle ranges and the first is hyphens ranges. These look like:

* `1.2 - 1.4.5` which is equivalent to `>= 1.2, <= 1.4.5`
* `2.3.4 - 4.5` which is equivalent to `>= 2.3.4, <= 4.5`

## Wildcards In Comparisons

The `x`, `X`, and `*` characters can be used as a wildcard character. This works for all comparison operators. When used on the `=` operator it falls back to the patch level comparison (see tilde below). For example,

* `1.2.x` is equivalent to `>= 1.2.0, < 1.3.0`
* `>= 1.2.x` is equivalent to `>= 1.2.0`
* `<= 2.x` is equivalent to `< 3`
* `*` is equivalent to `>= 0.0.0`

## Tilde Range Comparisons (Patch)

The tilde (`~`) comparison operator is for patch level ranges when a minor version is specified and major level changes when the minor number is missing. For example,

* `~1.2.3` is equivalent to `>= 1.2.3, < 1.3.0`
* `~1` is equivalent to `>= 1, < 2`
* `~2.3` is equivalent to `>= 2.3, < 2.4`
* `~1.2.x` is equivalent to `>= 1.2.0, < 1.3.0`
* `~1.x` is equivalent to `>= 1, < 2`

## Caret Range Comparisons (Major)

The caret (`^`) comparison operator is for major level changes. This is useful when comparisons of API versions as a major change is API breaking. For example,

* `^1.2.3` is equivalent to `>= 1.2.3, < 2.0.0`
* `^1.2.x` is equivalent to `>= 1.2.0, < 2.0.0`
* `^2.3` is equivalent to `>= 2.3, < 3`
* `^2.x` is equivalent to `>= 2.0.0, < 3`
