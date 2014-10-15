# Importing

Glide has limited support for importing from other formats.

**Note:** If you'd like to help build importers, we'd love some pull
requests. Just take a look at `cmd/godeps.git`.

## Godeps and Godeps-Git

To import from Godeps or Godeps-Git format, run `glide godeps`. This
will read the `glide.yaml`, then look for `Godeps` or `Godeps-Git` files
to also read. It will then attempt to merge the packages in those files
into the current YAML, printing the resulting YAML to standard out.

The preferred procedure for merging:

```
$ glide godeps # look at the output and see if it's okay
$ glide -q godeps > glide.yaml # Write the merged file
```
