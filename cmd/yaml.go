package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/yaml"
)

// YAML language compatible file extensions
var YAMLExtensions = []string{".yml", ".yaml"}

// ParseYaml parses the glide.yaml format and returns a Configuration object.
//
// Params:
//	- filename (string): YAML filename as a string
//
// Returns:
//	- *yaml.Config: The configuration.
func ParseYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)

	yml, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	cfg, err := yaml.FromYaml(string(yml))
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LookupManifestFile look up for a given file
// in the current working directory
func LookupManifestFile(fname string, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	ext := filepath.Ext(fname)
	if ext != ".yml" && ext != ".yaml" {
		return FileExists(fname)
	}

	var err cookoo.Interrupt
	var exists interface{}
	var name string = filepath.Base(fname[0 : len(fname)-len(ext)])

	for _, ext := range YAMLExtensions {
		filename := name + ext
		exists, err = FileExists(filename)
		if err != nil {
			continue
		}

		// Overwrite YAML file params
		params := p.AsMap()
		params["yaml"] = filename
		params["filename"] = filename
		p.Init(params)
		break
	}

	return exists, err
}

// FileExists verifies if the given filename exists in the
// current working directory
func FileExists(fname string) (interface{}, cookoo.Interrupt) {
	if _, err := os.Stat(fname); err != nil {
		cwd, _ := os.Getwd()
		return false, fmt.Errorf("%s is missing from %s", fname, cwd)
	}
	return true, nil
}

// ParseYamlString parses a YAML string. This is similar but different to
// ParseYaml that parses an external file.
//
// Params:
//	- yaml (string): YAML as a string.
//
// Returns:
//	- *yaml.Config: The configuration.
func ParseYamlString(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	yamlString := p.Get("yaml", "").(string)

	cfg, err := yaml.FromYaml(string(yamlString))
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// WriteYaml writes a yaml.Node to the console as a string.
//
// Params:
//	- conf: A *yaml.Config to render.
// 	- out (io.Writer): An output stream to write to. Default is os.Stdout.
// 	- filename (string): If set, the file will be opened and the content will be written to it.
func WriteYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*yaml.Config)
	toStdout := p.Get("toStdout", true).(bool)

	yml, err := yaml.ToYaml(cfg)
	if err != nil {
		return nil, err
	}
	var out io.Writer
	if nn, ok := p.Has("filename"); ok && len(nn.(string)) > 0 {
		file, err := os.Create(nn.(string))
		if err != nil {
		}
		defer file.Close()
		out = io.Writer(file)
		fmt.Fprint(out, yml)
	} else if toStdout {
		out = p.Get("out", os.Stdout).(io.Writer)
		fmt.Fprint(out, yml)
	}

	// Otherwise we supress output.
	return true, nil
}

// AddDependencies adds a list of *Dependency objects to the given *yaml.Config.
//
// This is used to merge in packages from other sources or config files.
func AddDependencies(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	deps := p.Get("dependencies", []*yaml.Dependency{}).([]*yaml.Dependency)
	config := p.Get("conf", nil).(*yaml.Config)

	// Make a set of existing package names for quick comparison.
	pkgSet := make(map[string]bool, len(config.Imports))
	for _, p := range config.Imports {
		pkgSet[p.Name] = true
	}

	// If a dep is not already present, add it.
	for _, dep := range deps {
		if _, ok := pkgSet[dep.Name]; ok {
			Warn("Package %s is already in glide.yaml. Skipping.\n", dep.Name)
			continue
		}
		config.Imports = append(config.Imports, dep)
	}

	return true, nil
}

// NormalizeName takes a package name and normalizes it to the top level package.
//
// For example, golang.org/x/crypto/ssh becomes golang.org/x/crypto. 'ssh' is
// returned as extra data.
func NormalizeName(name string) (string, string) {
	parts := strings.SplitN(name, "/", 4)
	extra := ""
	if len(parts) < 3 {
		return name, extra
	}
	if len(parts) == 4 {
		extra = parts[3]
	}
	return strings.Join(parts[0:3], "/"), extra
}
