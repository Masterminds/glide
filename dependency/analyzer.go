package dependency

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/gb"
	"github.com/Masterminds/glide/godep"
	"github.com/Masterminds/glide/gom"
	"github.com/Masterminds/glide/gpm"
	gpath "github.com/Masterminds/glide/path"
	"github.com/sdboyer/vsolver"
)

type notApplicable struct{}

func (notApplicable) Error() string {
	return ""
}

// Analyzer implements vsolver.ProjectAnalyzer. We inject the Analyzer into a
// vsolver.SourceManager, and it reports manifest and lock information to the
// SourceManager on request.
type Analyzer struct{}

func (a Analyzer) GetInfo(ctx build.Context, pn vsolver.ProjectName) (vsolver.Manifest, vsolver.Lock, error) {
	// For now, at least, we do not search above the root path provided by
	// the SourceManager.
	root := filepath.Join(ctx.GOPATH, string(pn))

	// this check should be unnecessary, but keeping it for now as a canary
	if _, err := os.Lstat(root); err != nil {
		return nil, nil, fmt.Errorf("No directory exists at %s; cannot produce ProjectInfo", root)
	}

	m, l, err := a.lookForGlide(root)
	if err == nil {
		// TODO verify project name is same as what SourceManager passed in?
		return m, l, nil
	} else if _, ok := err.(notApplicable); !ok {
		return nil, nil, err
	}

	// The happy path of finding both a glide manifest and lock file failed.
	// Now, we begin our descent, in which we attempt to divine exactly *which*
	// circle of hell we're in.

	// Try godep first
	m, l, err = a.lookForGodep(root)
	if err == nil {
		return m, l, nil
	} else if _, ok := err.(notApplicable); !ok {
		return nil, nil, err
	}

	// Next, gpm
	m, l, err = a.lookForGPM(root)
	if err == nil {
		return m, l, nil
	} else if _, ok := err.(notApplicable); !ok {
		return nil, nil, err
	}

	// Next, gb
	m, l, err = a.lookForGb(root)
	if err == nil {
		return m, l, nil
	} else if _, ok := err.(notApplicable); !ok {
		return nil, nil, err
	}

	// Next, gom
	m, l, err = a.lookForGom(root)
	if err == nil {
		return m, l, nil
	} else if _, ok := err.(notApplicable); !ok {
		return nil, nil, err
	}

	// TODO drop in glide's general analysis logic here

	return nil, nil, fmt.Errorf("No usable project data found")
}

func (a Analyzer) lookForGlide(root string) (vsolver.Manifest, vsolver.Lock, error) {
	mpath := filepath.Join(root, gpath.GlideFile)
	if _, err := os.Lstat(mpath); err != nil {
		return nil, nil, notApplicable{}
	}
	// Manifest found, so from here on, we're locked in - a returned error will
	// make it back to the SourceManager

	yml, err := ioutil.ReadFile(mpath)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while reading glide manifest data: %s", root)
	}

	m, err := cfg.ConfigFromYaml(yml)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while parsing glide manifest data: %s", root)
	}

	// Manifest found, read, and parsed - we're on the happy path. Whether we
	// find a lock or not, we will produce a valid result back to the
	// SourceManager.
	lpath := filepath.Join(root, gpath.LockFile)
	if _, err := os.Lstat(lpath); err != nil {
		return m, nil, nil
	}

	yml, err = ioutil.ReadFile(mpath)
	if err != nil {
		return m, nil, nil
	}

	l, err := cfg.LockfileFromYaml(yml)
	if err != nil {
		return m, nil, nil
	}

	return m, l, nil
}

func (a Analyzer) lookForGodep(root string) (vsolver.Manifest, vsolver.Lock, error) {
	if !godep.Has(root) {
		return nil, nil, notApplicable{}
	}

	d, l, err := godep.AsMetadataPair(root)
	if err != nil {
		return nil, nil, err
	}

	return &cfg.Config{ProjectName: root, Imports: d}, l, nil
}

func (a Analyzer) lookForGPM(root string) (vsolver.Manifest, vsolver.Lock, error) {
	if !gpm.Has(root) {
		return nil, nil, notApplicable{}
	}

	d, l, err := gpm.AsMetadataPair(root)
	if err != nil {
		return nil, nil, err
	}

	return &cfg.Config{ProjectName: root, Imports: d}, l, nil
}

func (a Analyzer) lookForGb(root string) (vsolver.Manifest, vsolver.Lock, error) {
	if !gpm.Has(root) {
		return nil, nil, notApplicable{}
	}

	d, l, err := gb.AsMetadataPair(root)
	if err != nil {
		return nil, nil, err
	}

	return &cfg.Config{ProjectName: root, Imports: d}, l, nil
}

func (a Analyzer) lookForGom(root string) (vsolver.Manifest, vsolver.Lock, error) {
	if !gpm.Has(root) {
		return nil, nil, notApplicable{}
	}

	return gom.AsMetadataPair(root)
}
