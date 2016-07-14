# Glide Plugins

Glide supports a simple plugin system similar to Git.

## Existing Plugins

Some plugins exist today for Glide including:

* [glide-vc](https://github.com/sgotti/glide-vc) - The vendor cleaner allows you to strip files files not needed for building your application from the `vendor/` directory.
* [glide-brew](https://github.com/heewa/glide-brew) - Convert Go deps managed by glide to Homebrew resources to help you make brew formulas for you Go programs.
* [glide-hash](https://github.com/mattfarina/glide-hash) - Generates a hash of the `glide.yaml` file compatible with Glides internal hash.

_Note, to add plugins to this list please create a pull request._

## How Plugins Work

When Glide encounters a subcommand that it does not know, it will try to delegate it to another executable according to the following rules.

Example:

```
$ glide install # We know this command, so we execute it
$ glide foo     # We don't know this command, so we look for a suitable
                # plugin.
```

In the example above, when glide receives the command `foo`, which it does not know, it will do the following:

1. Transform the name from `foo` to `glide-foo`
2. Look on the system `$PATH` for `glide-foo`. If it finds a program by that name, execute it...
3. Or else, look at the current project's root for `glide-foo`. (That is, look in the same directory as `glide.yaml`). If found, execute it.
4. If no suitable command is found, exit with an error.

## Writing a Glide Plugin

A Glide plugin can be written in any language you wish, provided that it can be executed from the command line as a subprocess of Glide. The example included with Glide is a simple Bash script. We could just as easily write Go, Python, Perl, or even Java code (with a wrapper) to
execute.

A glide plugin must be in one of two locations:

1. Somewhere on the PATH
2. In the same directory as `glide.yaml`

It is recommended that system-wide Glide plugins go in `/usr/local/bin` or `$GOPATH/bin` while project-specific plugins go in the same directory as `glide.yaml`.

### Arguments and Flags

Say Glide is executed like this:

```
$ glide foo -name=Matt myfile.txt
```

Glide will interpret this as a request to execute `glide-foo` with the arguments `-name=Matt myfile.txt`. It will not attempt to interpret those arguments or modify them in any way.

Hypothetically, if Glide had a `-x` flag of its own, you could call this:

```
$ glide -x foo -name=Matt myfile.txt
```

In this case, glide would interpret and swollow the -x and pass the rest on to `glide-foo` as in the example above.

## Example Plugin

File: glide-foo

```bash
#!/bin/bash

echo "Hello"
```
