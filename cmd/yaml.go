package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/cookoo"
	v "github.com/Masterminds/vcs"
	"github.com/kylelemons/go-gypsy/yaml"
)

// ParseYaml parses the glide.yaml format and returns a Configuration object.
//
// Params:
//	- filename (string): YAML filename as a string
//
// Context:
//	- yaml.File: This puts the parsed YAML file into the context.
//
// Returns:
//	- *Config: The configuration.
func ParseYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)
	//conf := new(Config)
	f, err := yaml.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	c.Put("yaml.File", f)
	return FromYaml(f.Root)
}

// ParseYamlString parses a YAML string. This is similar but different to
// ParseYaml that parses an external file.
//
// Params:
//	- yaml (string): YAML as a string.
//
// Returns:
//	- *Config: The configuration.
func ParseYamlString(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	yamlString := p.Get("yaml", "").(string)

	// Unfortunately, this does not wrap the root in a YAML file object.
	root, err := yaml.Parse(bytes.NewBufferString(yamlString))
	if err != nil {
		return nil, err
	}

	return FromYaml(root)
}

// WriteYaml writes a yaml.Node to the console as a string.
//
// Params:
//	- yaml.Node (yaml.Node): A yaml.Node to render.
// 	- out (io.Writer): An output stream to write to. Default is os.Stdout.
// 	- filename (string): If set, the file will be opened and the content will be written to it.
func WriteYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	top := p.Get("yaml.Node", yaml.Scalar("nothing to print")).(yaml.Node)
	var out io.Writer
	if nn, ok := p.Has("filename"); ok && len(nn.(string)) > 0 {
		file, err := os.Create(nn.(string))
		if err != nil {
		}
		defer file.Close()
		out = io.Writer(file)
	} else {
		out = p.Get("out", os.Stdout).(io.Writer)
	}

	fmt.Fprint(out, yaml.Render(top))

	return true, nil
}

// MergeToYaml converts a Config object and a yaml.File to a single yaml.File.
//
// Params:
//	- conf (*Config): The configuration to merge.
//	- overwriteImports (bool, default true): If this is true, old config will
//		overwritten. If false, we attempt to merge the old and new config, with
//		preference to the old.
//
// Returns:
//	- The root yaml.Node of the modified config.
//
// Uses:
//	- cxt.Get("yaml.File") as the source for the YAML file.
func MergeToYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	root := c.Get("yaml.File", nil).(*yaml.File).Root
	cfg := p.Get("conf", nil).(*Config)
	overwrite := p.Get("overwriteImports", true).(bool)

	rootMap, ok := root.(yaml.Map)
	if !ok {
		return nil, fmt.Errorf("Expected root node to be a map.")
	}

	if len(cfg.Name) > 0 {
		rootMap["package"] = yaml.Scalar(cfg.Name)
	}
	if cfg.InCommand != "" {
		rootMap["incmd"] = yaml.Scalar(cfg.InCommand)
	}

	if cfg.Flatten == true {
		rootMap["flatten"] = yaml.Scalar("true")
	}

	if overwrite {
		// Imports
		imports := make([]yaml.Node, len(cfg.Imports))
		for i, imp := range cfg.Imports {
			imports[i] = imp.ToYaml()
		}
		rootMap["import"] = yaml.List(imports)
	} else {
		var err error
		rootMap, err = mergeImports(rootMap, cfg)
		if err != nil {
			Warn("Problem merging imports: %s\n", err)
		}
	}

	return root, nil
}

// mergeImports merges the imports on a *Config into an existing YAML doc.
func mergeImports(root yaml.Map, cfg *Config) (yaml.Map, error) {
	left, err := FromYaml(root)
	if err != nil {
		return root, err
	}

	leftnames := make(map[string]bool, len(left.Imports))
	for _, i := range left.Imports {
		leftnames[i.Name] = true
	}

	for _, right := range cfg.Imports {
		if _, ok := leftnames[right.Name]; !ok {
			left.Imports = append(left.Imports, right)
		}
	}

	return left.ToYaml().(yaml.Map), nil
}

// AddDependencies adds a list of *Dependency objects to the given *Config.
//
// This is used to merge in packages from other sources or config files.
func AddDependencies(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	deps := p.Get("dependencies", []*Dependency{}).([]*Dependency)
	config := p.Get("conf", nil).(*Config)

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

func valOrEmpty(key string, store map[string]yaml.Node) string {
	val, ok := store[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(val.(yaml.Scalar).String())
}

// boolOrDefault returns a bool, with the dft returned if there is an error or the value is not true/false
func boolOrDefault(key string, store map[string]yaml.Node, dft bool) bool {
	val, ok := store[key]
	if !ok {
		return dft
	}
	switch val.(yaml.Scalar).String() {
	case "true":
		return true
	case "false":
		return false
	default:
		return dft
	}
}

// valOrList gets a single value or a list of values.
//
// Supports syntaxes like:
//
// 	subpkg: foo
//
// and
//
// 	supkpg:
// 		-foo
// 		-bar
func valOrList(key string, store map[string]yaml.Node) []string {
	val, ok := store[key]

	subpackages := []string{}
	if !ok {
		return subpackages
	}

	pkgs, ok := val.(yaml.List)

	if !ok {

		// Special case: Allow 'subpackages: justOne'
		if one, ok := val.(yaml.Scalar); ok {
			return []string{one.String()}
		}

		Warn("Expected list of subpackages.\n")
		return subpackages
	}

	for _, pkg := range pkgs {
		subpackages = append(subpackages, pkg.(yaml.Scalar).String())
	}
	return subpackages
}

func getVcsType(store map[string]yaml.Node) string {
	val, ok := store["vcs"]
	if !ok {
		return string(v.NoVCS)
	}

	name := val.(yaml.Scalar).String()

	switch name {
	case "git", "hg", "bzr", "svn":
		return name
	case "mercurial":
		return "hg"
	case "bazaar":
		return "bzr"
	case "subversion":
		return "svn"
	default:
		return ""
	}
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

// Config is the top-level configuration object.
type Config struct {
	Parent     *Config
	Name       string
	Imports    Dependencies
	DevImports Dependencies
	// InCommand is the default shell command run to start a 'glide in'
	// session.
	InCommand string
	Flatten   bool
}

// HasDependency returns true if the given name is listed as an import or dev import.
func (c *Config) HasDependency(name string) bool {
	for _, d := range c.Imports {
		if d.Name == name {
			return true
		}
	}
	for _, d := range c.DevImports {
		if d.Name == name {
			return true
		}
	}
	return false
}

// HasRecursiveDependency returns true if this config or one of it's parents has this dependency
func (c *Config) HasRecursiveDependency(name string) bool {
	if c.HasDependency(name) == true {
		return true
	} else if c.Parent != nil {
		return c.Parent.HasRecursiveDependency(name)
	}
	return false
}

// GetRoot follows the Parent down to the top node
func (c *Config) GetRoot() *Config {
	if c.Parent != nil {
		return c.Parent.GetRoot()
	}
	return c
}

// FromYaml creates a *Config from a  YAML node.
func FromYaml(top yaml.Node) (*Config, error) {
	conf := new(Config)

	vals, ok := top.(yaml.Map)
	if !ok {
		return conf, fmt.Errorf("Top YAML node must be a map.")
	}

	if name, ok := vals["package"]; ok {
		conf.Name = name.(yaml.Scalar).String()
	} else {
		Warn("The 'package' directive is required in Glide YAML.\n")
		conf.Name = "main"
	}

	// Allow the user to override the behavior of `glide in`.
	if incmd, ok := vals["incmd"]; ok {
		conf.InCommand = incmd.(yaml.Scalar).String()
	}

	// Package level Flatten
	conf.Flatten = boolOrDefault("flatten", vals, false)

	conf.Imports = make(Dependencies, 0, 1)
	if imp, ok := vals["import"]; ok {
		imports, ok := imp.(yaml.List)

		if ok {
			for _, v := range imports {
				dep, err := DependencyFromYaml(v)
				if err != nil {
					Warn("Could not add a dependency: %s\n", err)
				}
				conf.Imports = append(conf.Imports, dep)
			}
		}
	}

	// Same for (experimental) devimport.
	// These are currently unused. Not sure what we'll do with it yet.
	conf.DevImports = make(Dependencies, 0, 0)
	if imp, ok := vals["devimport"]; ok {
		imports, ok := imp.(yaml.List)
		if ok {
			for _, v := range imports {
				dep, err := DependencyFromYaml(v)
				if err != nil {
					Warn("Could not add a dependency: %s\n", err)
				}
				conf.DevImports = append(conf.DevImports, dep)
			}
		}
	}

	return conf, nil
}

// ToYaml returns a yaml.Map containing the data from Config.
func (c *Config) ToYaml() yaml.Node {
	cfg := make(map[string]yaml.Node, 5)

	cfg["package"] = yaml.Scalar(c.Name)
	if len(c.InCommand) > 0 {
		cfg["incmd"] = yaml.Scalar(c.InCommand)
	}
	if c.Flatten == true {
		cfg["flatten"] = yaml.Scalar("true")
	}

	imps := make([]yaml.Node, len(c.Imports))
	for i, imp := range c.Imports {
		imps[i] = imp.ToYaml()
	}
	devimps := make([]yaml.Node, len(c.DevImports))
	for i, dimp := range c.DevImports {
		devimps[i] = dimp.ToYaml()
	}

	// Fixed in 0.5.0. Prior to that, these were not being printed. Worried
	// that the "fix" might introduce an unintended side effect.
	if len(imps) > 0 {
		cfg["import"] = yaml.List(imps)
	}
	if len(devimps) > 0 {
		cfg["devimport"] = yaml.List(devimps)
	}

	return yaml.Map(cfg)
}

// Dependency describes a package that the present package depends upon.
type Dependency struct {
	Name, Reference, Repository string
	VcsType                     string
	Subpackages, Arch, Os       []string
	UpdateAsVendored            bool
	Flatten                     bool
	Flattened                   bool
}

// DependencyFromYaml creates a dependency from a yaml.Node.
func DependencyFromYaml(node yaml.Node) (*Dependency, error) {
	pkg, ok := node.(yaml.Map)
	if !ok {
		return &Dependency{}, fmt.Errorf("Expected yaml.Node to be a dependency map.")
	}
	dep := &Dependency{
		Name:        valOrEmpty("package", pkg),
		Reference:   valOrEmpty("ref", pkg),
		VcsType:     getVcsType(pkg),
		Repository:  valOrEmpty("repo", pkg),
		Subpackages: valOrList("subpackages", pkg),
		Arch:        valOrList("arch", pkg),
		Os:          valOrList("os", pkg),
		Flatten:     boolOrDefault("flatten", pkg, false),
	}

	return dep, nil
}

// GetRepo retrieves a Masterminds/vcs repo object configured for the root
// of the package being retrieved.
func (d *Dependency) GetRepo(dest string) (v.Repo, error) {

	// The remote location is either the configured repo or the package
	// name as an https url.
	var remote string
	if len(d.Repository) > 0 {
		remote = d.Repository
	} else {
		remote = "https://" + d.Name
	}

	// If the VCS type has a value we try that first.
	if len(d.VcsType) > 0 && d.VcsType != "None" {
		switch v.Type(d.VcsType) {
		case v.Git:
			return v.NewGitRepo(remote, dest)
		case v.Svn:
			return v.NewSvnRepo(remote, dest)
		case v.Hg:
			return v.NewHgRepo(remote, dest)
		case v.Bzr:
			return v.NewBzrRepo(remote, dest)
		default:
			return nil, fmt.Errorf("Unknown VCS type %s set for %s", d.VcsType, d.Name)
		}
	}

	// When no type set we try to autodetect.
	return v.NewRepo(remote, dest)
}

func stripScheme(u string) string {
	parts := strings.Split(u, "://")
	if len(parts) > 1 {
		return parts[1]
	}
	return u
}

// ToYaml converts a *Dependency to a YAML Map node.
func (d *Dependency) ToYaml() yaml.Node {
	dep := make(map[string]yaml.Node, 8)
	dep["package"] = yaml.Scalar(d.Name)

	if len(d.Subpackages) > 0 {
		subp := make([]yaml.Node, len(d.Subpackages))
		for i, item := range d.Subpackages {
			subp[i] = yaml.Scalar(item)
		}

		dep["subpackages"] = yaml.List(subp)
	}
	vcs := d.VcsType
	if len(vcs) > 0 {
		dep["vcs"] = yaml.Scalar(vcs)
	}
	if len(d.Reference) > 0 {
		dep["ref"] = yaml.Scalar(d.Reference)
	}
	if len(d.Repository) > 0 {
		dep["repo"] = yaml.Scalar(d.Repository)
	}

	if len(d.Arch) > 0 {
		archs := make([]yaml.Node, len(d.Arch))
		for i, a := range d.Arch {
			archs[i] = yaml.Scalar(a)
		}
		dep["arch"] = yaml.List(archs)
	}
	if len(d.Os) > 0 {
		oses := make([]yaml.Node, len(d.Os))
		for i, a := range d.Os {
			oses[i] = yaml.Scalar(a)
		}
		dep["os"] = yaml.List(oses)
	}

	// Note, the yaml package we use sorts strings of scalars so flatten
	// will always be the top item.
	if d.Flatten == true {
		dep["flatten"] = yaml.Scalar("true")
	}

	return yaml.Map(dep)
}

// Dependencies is a collection of Dependency
type Dependencies []*Dependency

// Get a dependency by name
func (d Dependencies) Get(name string) *Dependency {
	for _, dep := range d {
		if dep.Name == name {
			return dep
		}
	}
	return nil
}
