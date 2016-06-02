package action

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/gb"
	"github.com/Masterminds/glide/godep"
	"github.com/Masterminds/glide/gpm"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// Create creates/initializes a new Glide repository.
//
// This will fail if a glide.yaml already exists.
//
// By default, this will scan the present source code directory for dependencies.
//
// If skipImport is set to true, this will not attempt to import from an existing
// GPM, Godep, or GB project if one should exist. However, it will still attempt
// to read the local source to determine required packages.
func Create(base string, skipImport, nonInteractive bool) {
	glidefile := gpath.GlideFile
	// Guard against overwrites.
	guardYAML(glidefile)

	// Guess deps
	conf := guessDeps(base, skipImport)
	// Write YAML
	msg.Info("Writing configuration file (%s)", glidefile)
	if err := conf.WriteFile(glidefile); err != nil {
		msg.Die("Could not save %s: %s", glidefile, err)
	}

	var res bool
	if !nonInteractive {
		msg.Info("Would you like Glide to help you find ways to improve your glide.yaml configuration?")
		msg.Info("If you want to revisit this step you can use the config-wizard command at any time.")
		msg.Info("Yes (Y) or No (N)?")
		res = msg.PromptUntilYorN()
		if res {
			ConfigWizard(base)
		}
	}

	if !res {
		msg.Info("You can now edit the glide.yaml file. Consider:")
		msg.Info("--> Using versions and ranges. See https://glide.sh/docs/versions/")
		msg.Info("--> Adding additional metadata. See https://glide.sh/docs/glide.yaml/")
		msg.Info("--> Running the config-wizard command to improve the versions in your configuration")
	}
}

// guardYAML fails if the given file already exists.
//
// This prevents an important file from being overwritten.
func guardYAML(filename string) {
	if _, err := os.Stat(filename); err == nil {
		msg.Die("Cowardly refusing to overwrite existing YAML.")
	}
}

// guessDeps attempts to resolve all of the dependencies for a given project.
//
// base is the directory to start with.
// skipImport will skip running the automatic imports.
//
// FIXME: This function is likely a one-off that has a more standard alternative.
// It's also long and could use a refactor.
func guessDeps(base string, skipImport bool) *cfg.Config {
	buildContext, err := util.GetBuildContext()
	if err != nil {
		msg.Die("Failed to build an import context: %s", err)
	}
	name := buildContext.PackageName(base)

	msg.Info("Generating a YAML configuration file and guessing the dependencies")

	config := new(cfg.Config)

	// Get the name of the top level package
	config.Name = name

	// Import by looking at other package managers and looking over the
	// entire directory structure.

	// Attempt to import from other package managers.
	if !skipImport {
		guessImportDeps(base, config)
	}

	importLen := len(config.Imports)
	if importLen == 0 {
		msg.Info("Scanning code to look for dependencies")
	} else {
		msg.Info("Scanning code to look for dependencies not found in import")
	}

	// Resolve dependencies by looking at the tree.
	r, err := dependency.NewResolver(base)
	if err != nil {
		msg.Die("Error creating a dependency resolver: %s", err)
	}

	h := &dependency.DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}}
	r.Handler = h

	sortable, err := r.ResolveLocal(false)
	if err != nil {
		msg.Die("Error resolving local dependencies: %s", err)
	}

	sort.Strings(sortable)

	vpath := r.VendorDir
	if !strings.HasSuffix(vpath, "/") {
		vpath = vpath + string(os.PathSeparator)
	}

	for _, pa := range sortable {
		n := strings.TrimPrefix(pa, vpath)
		root, subpkg := util.NormalizeName(n)

		if !config.HasDependency(root) && root != config.Name {
			msg.Info("--> Found reference to %s\n", n)
			d := &cfg.Dependency{
				Name: root,
			}
			if len(subpkg) > 0 {
				d.Subpackages = []string{subpkg}
			}
			config.Imports = append(config.Imports, d)
		} else if config.HasDependency(root) {
			if len(subpkg) > 0 {
				subpkg = strings.TrimPrefix(subpkg, "/")
				d := config.Imports.Get(root)
				if !d.HasSubpackage(subpkg) {
					msg.Info("--> Adding sub-package %s to %s\n", subpkg, root)
					d.Subpackages = append(d.Subpackages, subpkg)
				}
			}
		}
	}

	if len(config.Imports) == importLen && importLen != 0 {
		msg.Info("--> Code scanning found no additional imports")
	}

	return config
}

func guessImportDeps(base string, config *cfg.Config) {
	msg.Info("Attempting to import from other package managers (use --skip-import to skip)")
	deps := []*cfg.Dependency{}
	absBase, err := filepath.Abs(base)
	if err != nil {
		msg.Die("Failed to resolve location of %s: %s", base, err)
	}

	if d, ok := guessImportGodep(absBase); ok {
		msg.Info("Importing Godep configuration")
		msg.Warn("Godep uses commit id versions. Consider using Semantic Versions with Glide")
		deps = d
	} else if d, ok := guessImportGPM(absBase); ok {
		msg.Info("Importing GPM configuration")
		deps = d
	} else if d, ok := guessImportGB(absBase); ok {
		msg.Info("Importing GB configuration")
		deps = d
	}

	for _, i := range deps {
		if i.Reference == "" {
			msg.Info("--> Found imported reference to %s", i.Name)
		} else {
			msg.Info("--> Found imported reference to %s at revision %s", i.Name, i.Reference)
		}

		config.Imports = append(config.Imports, i)
	}
}

func guessImportGodep(dir string) ([]*cfg.Dependency, bool) {
	d, err := godep.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGPM(dir string) ([]*cfg.Dependency, bool) {
	d, err := gpm.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGB(dir string) ([]*cfg.Dependency, bool) {
	d, err := gb.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}
