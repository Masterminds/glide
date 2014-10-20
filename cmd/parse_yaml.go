package cmd

import (
	"fmt"
	"strings"

	"github.com/Masterminds/cookoo"
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

// WriteYaml writes a yaml.Node to the console as a string.
//
// Params:
//	- yaml.Node: A yaml.Node to render.
func WriteYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	top := p.Get("yaml.Node", yaml.Scalar("nothing to print")).(yaml.Node)

	fmt.Print(yaml.Render(top))

	return true, nil
}

// Convert a Config object and a yaml.File to a single yaml.File.
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

func vcsString(vtype uint) string {
	switch vtype {
	case Git:
		return "git"
	case Hg:
		return "hg"
	case Bzr:
		return "bzr"
	case Svn:
		return "svn"
	default:
		return ""
	}
}

func valOrEmpty(key string, store map[string]yaml.Node) string {
	val, ok := store[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(val.(yaml.Scalar).String())
}

func subpkg(key string, store map[string]yaml.Node) []string {
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

func getVcsType(store map[string]yaml.Node) uint {
	val, ok := store["vcs"]
	if !ok {
		return NoVCS
	}

	name := val.(yaml.Scalar).String()

	switch name {
	case "git":
		return Git
	case "hg", "mercurial":
		return Hg
	case "bzr", "bazaar":
		return Bzr
	case "svn", "subversion":
		return Svn
	default:
		return NoVCS
	}
}

// Config is the top-level configuration object.
type Config struct {
	Name       string
	Imports    []*Dependency
	DevImports []*Dependency
	// InCommand is the default shell command run to start a 'glide in'
	// session.
	InCommand string
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

	conf.Imports = make([]*Dependency, 0, 1)
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
	conf.DevImports = []*Dependency{}
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
	cfg := make(map[string]yaml.Node, 4)

	cfg["package"] = yaml.Scalar(c.Name)
	if len(c.InCommand) > 0 {
		cfg["incmd"] = yaml.Scalar(c.InCommand)
	}

	imps := make([]yaml.Node, len(c.Imports))
	for i, imp := range c.Imports {
		imps[i] = imp.ToYaml()
	}
	devimps := make([]yaml.Node, len(c.DevImports))
	for i, dimp := range c.DevImports {
		devimps[i] = dimp.ToYaml()
	}

	return yaml.Map(cfg)
}

// Dependency describes a package that the present package depends upon.
type Dependency struct {
	Name, Reference, Repository string
	VcsType                     uint
	Subpackages                 []string
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
		Subpackages: subpkg("subpackages", pkg),
	}

	return dep, nil
}

// ToYaml converts a *Dependency to a YAML Map node.
func (d *Dependency) ToYaml() yaml.Node {
	dep := make(map[string]yaml.Node, 5)
	dep["package"] = yaml.Scalar(d.Name)

	subp := make([]yaml.Node, len(d.Subpackages))
	for i, item := range d.Subpackages {
		subp[i] = yaml.Scalar(item)
	}

	dep["subpackages"] = yaml.List(subp)
	vcs := vcsString(d.VcsType)
	if len(vcs) > 0 {
		dep["vcs"] = yaml.Scalar(vcs)
	}
	if len(d.Reference) > 0 {
		dep["ref"] = yaml.Scalar(d.Reference)
	}
	if len(d.Repository) > 0 {
		dep["repo"] = yaml.Scalar(d.Repository)
	}
	return yaml.Map(dep)
}
