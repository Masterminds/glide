package gps

import (
	"fmt"
	"path/filepath"
)

// dsp - "depspec with packages"
//
// Wraps a set of tpkgs onto a depspec, and returns it.
func dsp(ds depspec, pkgs ...tpkg) depspec {
	ds.pkgs = pkgs
	return ds
}

// pkg makes a tpkg appropriate for use in bimodal testing
func pkg(path string, imports ...string) tpkg {
	return tpkg{
		path:    path,
		imports: imports,
	}
}

func init() {
	for k, fix := range bimodalFixtures {
		// Assign the name into the fixture itself
		fix.n = k
		bimodalFixtures[k] = fix
	}
}

// Fixtures that rely on simulated bimodal (project and package-level)
// analysis for correct operation. The name given in the map gets assigned into
// the fixture itself in init().
var bimodalFixtures = map[string]bimodalFixture{
	// Simple case, ensures that we do the very basics of picking up and
	// including a single, simple import that is not expressed as a constraint
	"simple bm-add": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "a")),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a")),
		},
		r: mksolution(
			"a 1.0.0",
		),
	},
	// Ensure it works when the import jump is not from the package with the
	// same path as root, but from a subpkg
	"subpkg bm-add": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a"),
			),
		},
		r: mksolution(
			"a 1.0.0",
		),
	},
	// The same, but with a jump through two subpkgs
	"double-subpkg bm-add": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "root/bar"),
				pkg("root/bar", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a"),
			),
		},
		r: mksolution(
			"a 1.0.0",
		),
	},
	// Same again, but now nest the subpkgs
	"double nested subpkg bm-add": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "root/foo/bar"),
				pkg("root/foo/bar", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a"),
			),
		},
		r: mksolution(
			"a 1.0.0",
		),
	},
	// Importing package from project with no root package
	"bm-add on project with no pkg in root dir": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "a/foo")),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a/foo")),
		},
		r: mksolution(
			"a 1.0.0",
		),
	},
	// Import jump is in a dep, and points to a transitive dep
	"transitive bm-add": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "b"),
			),
			dsp(mkDepspec("b 1.0.0"),
				pkg("b"),
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 1.0.0",
		),
	},
	// Constraints apply only if the project that declares them has a
	// reachable import
	"constraints activated by import": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0", "b 1.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "b"),
			),
			dsp(mkDepspec("b 1.0.0"),
				pkg("b"),
			),
			dsp(mkDepspec("b 1.1.0"),
				pkg("b"),
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 1.1.0",
		),
	},
	// Import jump is in a dep, and points to a transitive dep - but only in not
	// the first version we try
	"transitive bm-add on older version": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0", "a ~1.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "b"),
			),
			dsp(mkDepspec("a 1.1.0"),
				pkg("a"),
			),
			dsp(mkDepspec("b 1.0.0"),
				pkg("b"),
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 1.0.0",
		),
	},
	// Import jump is in a dep, and points to a transitive dep - but will only
	// get there via backtracking
	"backtrack to dep on bm-add": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "a", "b"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "c"),
			),
			dsp(mkDepspec("a 1.1.0"),
				pkg("a"),
			),
			// Include two versions of b, otherwise it'll be selected first
			dsp(mkDepspec("b 0.9.0"),
				pkg("b", "c"),
			),
			dsp(mkDepspec("b 1.0.0"),
				pkg("b", "c"),
			),
			dsp(mkDepspec("c 1.0.0", "a 1.0.0"),
				pkg("c", "a"),
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 1.0.0",
			"c 1.0.0",
		),
	},
	// Import jump is in a dep subpkg, and points to a transitive dep
	"transitive subpkg bm-add": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "a/bar"),
				pkg("a/bar", "b"),
			),
			dsp(mkDepspec("b 1.0.0"),
				pkg("b"),
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 1.0.0",
		),
	},
	// Import jump is in a dep subpkg, pointing to a transitive dep, but only in
	// not the first version we try
	"transitive subpkg bm-add on older version": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0", "a ~1.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "a/bar"),
				pkg("a/bar", "b"),
			),
			dsp(mkDepspec("a 1.1.0"),
				pkg("a", "a/bar"),
				pkg("a/bar"),
			),
			dsp(mkDepspec("b 1.0.0"),
				pkg("b"),
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 1.0.0",
		),
	},
	// Ensure that if a constraint is expressed, but no actual import exists,
	// then the constraint is disregarded - the project named in the constraint
	// is not part of the solution.
	"ignore constraint without import": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0", "a 1.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a"),
			),
		},
		r: mksolution(),
	},
	// Transitive deps from one project (a) get incrementally included as other
	// deps incorporate its various packages.
	"multi-stage pkg incorporation": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "a", "d"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "b"),
				pkg("a/second", "c"),
			),
			dsp(mkDepspec("b 2.0.0"),
				pkg("b"),
			),
			dsp(mkDepspec("c 1.2.0"),
				pkg("c"),
			),
			dsp(mkDepspec("d 1.0.0"),
				pkg("d", "a/second"),
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 2.0.0",
			"c 1.2.0",
			"d 1.0.0",
		),
	},
	// Regression - make sure that the the constraint/import intersector only
	// accepts a project 'match' if exactly equal, or a separating slash is
	// present.
	"radix path separator post-check": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "foo", "foobar"),
			),
			dsp(mkDepspec("foo 1.0.0"),
				pkg("foo"),
			),
			dsp(mkDepspec("foobar 1.0.0"),
				pkg("foobar"),
			),
		},
		r: mksolution(
			"foo 1.0.0",
			"foobar 1.0.0",
		),
	},
	// Well-formed failure when there's a dependency on a pkg that doesn't exist
	"fail when imports nonexistent package": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0", "a 1.0.0"),
				pkg("root", "a/foo"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a"),
			),
		},
		fail: &noVersionError{
			pn: mkPI("a"),
			fails: []failedVersion{
				{
					v: NewVersion("1.0.0"),
					f: &checkeeHasProblemPackagesFailure{
						goal: mkAtom("a 1.0.0"),
						failpkg: map[string]errDeppers{
							"a/foo": errDeppers{
								err: nil, // nil indicates package is missing
								deppers: []atom{
									mkAtom("root"),
								},
							},
						},
					},
				},
			},
		},
	},
	// Transitive deps from one project (a) get incrementally included as other
	// deps incorporate its various packages, and fail with proper error when we
	// discover one incrementally that isn't present
	"fail multi-stage missing pkg": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "a", "d"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "b"),
				pkg("a/second", "c"),
			),
			dsp(mkDepspec("b 2.0.0"),
				pkg("b"),
			),
			dsp(mkDepspec("c 1.2.0"),
				pkg("c"),
			),
			dsp(mkDepspec("d 1.0.0"),
				pkg("d", "a/second"),
				pkg("d", "a/nonexistent"),
			),
		},
		fail: &noVersionError{
			pn: mkPI("d"),
			fails: []failedVersion{
				{
					v: NewVersion("1.0.0"),
					f: &depHasProblemPackagesFailure{
						goal: mkADep("d 1.0.0", "a", Any(), "a/nonexistent"),
						v:    NewVersion("1.0.0"),
						prob: map[string]error{
							"a/nonexistent": nil,
						},
					},
				},
			},
		},
	},
	// Check ignores on the root project
	"ignore in double-subpkg": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "root/bar", "b"),
				pkg("root/bar", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a"),
			),
			dsp(mkDepspec("b 1.0.0"),
				pkg("b"),
			),
		},
		ignore: []string{"root/bar"},
		r: mksolution(
			"b 1.0.0",
		),
	},
	// Ignores on a dep pkg
	"ignore through dep pkg": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "root/foo"),
				pkg("root/foo", "a"),
			),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "a/bar"),
				pkg("a/bar", "b"),
			),
			dsp(mkDepspec("b 1.0.0"),
				pkg("b"),
			),
		},
		ignore: []string{"a/bar"},
		r: mksolution(
			"a 1.0.0",
		),
	},
	// Preferred version, as derived from a dep's lock, is attempted first
	"respect prefv, simple case": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "a")),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "b")),
			dsp(mkDepspec("b 1.0.0 foorev"),
				pkg("b")),
			dsp(mkDepspec("b 2.0.0 barrev"),
				pkg("b")),
		},
		lm: map[string]fixLock{
			"a 1.0.0": mklock(
				"b 1.0.0 foorev",
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 1.0.0 foorev",
		),
	},
	// Preferred version, as derived from a dep's lock, is attempted first, even
	// if the root also has a direct dep on it (root doesn't need to use
	// preferreds, because it has direct control AND because the root lock
	// already supercedes dep lock "preferences")
	"respect dep prefv with root import": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "a", "b")),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "b")),
			//dsp(newDepspec("a 1.0.1"),
			//pkg("a", "b")),
			//dsp(newDepspec("a 1.1.0"),
			//pkg("a", "b")),
			dsp(mkDepspec("b 1.0.0 foorev"),
				pkg("b")),
			dsp(mkDepspec("b 2.0.0 barrev"),
				pkg("b")),
		},
		lm: map[string]fixLock{
			"a 1.0.0": mklock(
				"b 1.0.0 foorev",
			),
		},
		r: mksolution(
			"a 1.0.0",
			"b 1.0.0 foorev",
		),
	},
	// Preferred versions can only work if the thing offering it has been
	// selected, or at least marked in the unselected queue
	"prefv only works if depper is selected": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "a", "b")),
			// Three atoms for a, which will mean it gets visited after b
			dsp(mkDepspec("a 1.0.0"),
				pkg("a", "b")),
			dsp(mkDepspec("a 1.0.1"),
				pkg("a", "b")),
			dsp(mkDepspec("a 1.1.0"),
				pkg("a", "b")),
			dsp(mkDepspec("b 1.0.0 foorev"),
				pkg("b")),
			dsp(mkDepspec("b 2.0.0 barrev"),
				pkg("b")),
		},
		lm: map[string]fixLock{
			"a 1.0.0": mklock(
				"b 1.0.0 foorev",
			),
		},
		r: mksolution(
			"a 1.1.0",
			"b 2.0.0 barrev",
		),
	},
	"override unconstrained root import": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "a")),
			dsp(mkDepspec("a 1.0.0"),
				pkg("a")),
			dsp(mkDepspec("a 2.0.0"),
				pkg("a")),
		},
		ovr: ProjectConstraints{
			ProjectRoot("a"): ProjectProperties{
				Constraint: NewVersion("1.0.0"),
			},
		},
		r: mksolution(
			"a 1.0.0",
		),
	},
	"overridden mismatched net addrs, alt in dep": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0"),
				pkg("root", "foo")),
			dsp(mkDepspec("foo 1.0.0", "bar from baz 1.0.0"),
				pkg("foo", "bar")),
			dsp(mkDepspec("bar 1.0.0"),
				pkg("bar")),
			dsp(mkDepspec("baz 1.0.0"),
				pkg("bar")),
		},
		ovr: ProjectConstraints{
			ProjectRoot("bar"): ProjectProperties{
				NetworkName: "baz",
			},
		},
		r: mksolution(
			"foo 1.0.0",
			"bar from baz 1.0.0",
		),
	},
	"overridden mismatched net addrs, alt in root": {
		ds: []depspec{
			dsp(mkDepspec("root 0.0.0", "bar from baz 1.0.0"),
				pkg("root", "foo")),
			dsp(mkDepspec("foo 1.0.0"),
				pkg("foo", "bar")),
			dsp(mkDepspec("bar 1.0.0"),
				pkg("bar")),
			dsp(mkDepspec("baz 1.0.0"),
				pkg("bar")),
		},
		ovr: ProjectConstraints{
			ProjectRoot("bar"): ProjectProperties{
				NetworkName: "baz",
			},
		},
		r: mksolution(
			"foo 1.0.0",
			"bar from baz 1.0.0",
		),
	},
}

// tpkg is a representation of a single package. It has its own import path, as
// well as a list of paths it itself "imports".
type tpkg struct {
	// Full import path of this package
	path string
	// Slice of full paths to its virtual imports
	imports []string
}

type bimodalFixture struct {
	// name of this fixture datum
	n string
	// bimodal project. first is always treated as root project
	ds []depspec
	// results; map of name/version pairs
	r map[string]Version
	// max attempts the solver should need to find solution. 0 means no limit
	maxAttempts int
	// Use downgrade instead of default upgrade sorter
	downgrade bool
	// lock file simulator, if one's to be used at all
	l fixLock
	// map of locks for deps, if any. keys should be of the form:
	// "<project> <version>"
	lm map[string]fixLock
	// solve failure expected, if any
	fail error
	// overrides, if any
	ovr ProjectConstraints
	// request up/downgrade to all projects
	changeall bool
	// pkgs to ignore
	ignore []string
}

func (f bimodalFixture) name() string {
	return f.n
}

func (f bimodalFixture) specs() []depspec {
	return f.ds
}

func (f bimodalFixture) maxTries() int {
	return f.maxAttempts
}

func (f bimodalFixture) solution() map[string]Version {
	return f.r
}

func (f bimodalFixture) rootmanifest() RootManifest {
	m := simpleRootManifest{
		c:   f.ds[0].deps,
		tc:  f.ds[0].devdeps,
		ovr: f.ovr,
		ig:  make(map[string]bool),
	}
	for _, ig := range f.ignore {
		m.ig[ig] = true
	}

	return m
}

func (f bimodalFixture) failure() error {
	return f.fail
}

// bmSourceManager is an SM specifically for the bimodal fixtures. It composes
// the general depspec SM, and differs from it in how it answers static analysis
// calls, and its support for package ignores and dep lock data.
type bmSourceManager struct {
	depspecSourceManager
	lm map[string]fixLock
}

var _ SourceManager = &bmSourceManager{}

func newbmSM(bmf bimodalFixture) *bmSourceManager {
	sm := &bmSourceManager{
		depspecSourceManager: *newdepspecSM(bmf.ds, bmf.ignore),
	}
	sm.rm = computeBimodalExternalMap(bmf.ds)
	sm.lm = bmf.lm

	return sm
}

func (sm *bmSourceManager) ListPackages(id ProjectIdentifier, v Version) (PackageTree, error) {
	for k, ds := range sm.specs {
		// Cheat for root, otherwise we blow up b/c version is empty
		if id.ProjectRoot == ds.n && (k == 0 || ds.v.Matches(v)) {
			ptree := PackageTree{
				ImportRoot: string(id.ProjectRoot),
				Packages:   make(map[string]PackageOrErr),
			}
			for _, pkg := range ds.pkgs {
				ptree.Packages[pkg.path] = PackageOrErr{
					P: Package{
						ImportPath: pkg.path,
						Name:       filepath.Base(pkg.path),
						Imports:    pkg.imports,
					},
				}
			}

			return ptree, nil
		}
	}

	return PackageTree{}, fmt.Errorf("Project %s at version %s could not be found", id.errString(), v)
}

func (sm *bmSourceManager) GetManifestAndLock(id ProjectIdentifier, v Version) (Manifest, Lock, error) {
	for _, ds := range sm.specs {
		if id.ProjectRoot == ds.n && v.Matches(ds.v) {
			if l, exists := sm.lm[string(id.ProjectRoot)+" "+v.String()]; exists {
				return ds, l, nil
			}
			return ds, dummyLock{}, nil
		}
	}

	// TODO(sdboyer) proper solver-type errors
	return nil, nil, fmt.Errorf("Project %s at version %s could not be found", id.errString(), v)
}

// computeBimodalExternalMap takes a set of depspecs and computes an
// internally-versioned external reach map that is useful for quickly answering
// ListExternal()-type calls.
//
// Note that it does not do things like stripping out stdlib packages - these
// maps are intended for use in SM fixtures, and that's a higher-level
// responsibility within the system.
func computeBimodalExternalMap(ds []depspec) map[pident]map[string][]string {
	// map of project name+version -> map of subpkg name -> external pkg list
	rm := make(map[pident]map[string][]string)

	// algorithm adapted from externalReach()
	for _, d := range ds {
		// Keeps a list of all internal and external reaches for packages within
		// a given root. We create one on each pass through, rather than doing
		// them all at once, because the depspec set may (read: is expected to)
		// have multiple versions of the same base project, and each of those
		// must be calculated independently.
		workmap := make(map[string]wm)

		for _, pkg := range d.pkgs {
			w := wm{
				ex: make(map[string]bool),
				in: make(map[string]bool),
			}

			for _, imp := range pkg.imports {
				if !checkPrefixSlash(filepath.Clean(imp), string(d.n)) {
					// Easy case - if the import is not a child of the base
					// project path, put it in the external map
					w.ex[imp] = true
				} else {
					if w2, seen := workmap[imp]; seen {
						// If it is, and we've seen that path, dereference it
						// immediately
						for i := range w2.ex {
							w.ex[i] = true
						}
						for i := range w2.in {
							w.in[i] = true
						}
					} else {
						// Otherwise, put it in the 'in' map for later
						// reprocessing
						w.in[imp] = true
					}
				}
			}
			workmap[pkg.path] = w
		}

		drm := wmToReach(workmap, "")
		rm[pident{n: d.n, v: d.v}] = drm
	}

	return rm
}
