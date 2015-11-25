package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
)

// ParseYaml parses the glide.yaml format and returns a Configuration object.
//
// Params:
//	- filename (string): YAML filename as a string
//
// Returns:
//	- *cfg.Config: The configuration.
func ParseYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)
	//conf := new(Config)
	yml, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	conf, err := cfg.ConfigFromYaml(yml)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// ParseYamlString parses a YAML string. This is similar but different to
// ParseYaml that parses an external file.
//
// Params:
//	- yaml (string): YAML as a string.
//
// Returns:
//	- *cfg.Config: The configuration.
func ParseYamlString(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	yamlString := p.Get("yaml", "").(string)

	conf, err := cfg.ConfigFromYaml([]byte(yamlString))
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// GuardYaml protects the glide yaml file from being overwritten.
func GuardYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)
	if _, err := os.Stat(fname); err == nil {
		cwd, _ := os.Getwd()
		return false, fmt.Errorf("Cowardly refusing to overwrite %s in %s", fname, cwd)
	}

	return true, nil
}

// WriteYaml writes the config as YAML.
//
// Params:
//	- conf: A *cfg.Config to render.
// 	- out (io.Writer): An output stream to write to. Default is os.Stdout.
// 	- filename (string): If set, the file will be opened and the content will be written to it.
func WriteYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	conf := p.Get("conf", nil).(*cfg.Config)
	toStdout := p.Get("toStdout", true).(bool)

	data, err := conf.Marshal()
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
		//fmt.Fprint(out, yml)
		out.Write(data)
	} else if toStdout {
		out = p.Get("out", os.Stdout).(io.Writer)
		//fmt.Fprint(out, yml)
		out.Write(data)
	}

	// Otherwise we supress output.
	return true, nil
}

// WriteLock writes the lock as YAML.
//
// Params:
//	- lockfile: A *cfg.Lockfile to render.
// 	- out (io.Writer): An output stream to write to. Default is os.Stdout.
func WriteLock(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	lockfile := p.Get("lockfile", nil).(*cfg.Lockfile)

	data, err := lockfile.Marshal()
	if err != nil {
		return nil, err
	}

	var out io.Writer
	file, err := os.Create("glide.lock")
	if err != nil {
		return false, err
	}
	defer file.Close()
	out = io.Writer(file)
	out.Write(data)

	return true, nil
}

// AddDependencies adds a list of *Dependency objects to the given *cfg.Config.
//
// This is used to merge in packages from other sources or config files.
func AddDependencies(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	deps := p.Get("dependencies", []*cfg.Dependency{}).([]*cfg.Dependency)
	config := p.Get("conf", nil).(*cfg.Config)

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
