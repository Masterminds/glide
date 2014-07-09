package main

import (
	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/cli"
	"github.com/kylelemons/go-gypsy/yaml"

	"path"
	"flag"
	"fmt"
	"os/exec"
	"os"
)

const Summary = "Manage Go projects with ease."
const Usage = `Manage dependencies, naming, and GOPATH for your Go projects.

Examples:
	$ source glide in
	$ glide init
	$ glide install

COMMANDS
========

- help: Show this help message (alias of -h)
- in: Set the GOPATH. Usage: "source glide in"
- init: Initialize a new project
- install: Install all packages in the glide.yaml
- update: Update existing packages
- prebuild: Prebuild libraries into .a files.
- clean: Remove prebuilt libraries.

`

func main() {
	reg, router, cxt := cookoo.Cookoo()

	routes(reg, cxt)

	if err := router.HandleRequest("@startup", cxt, false); err != nil {
		fmt.Printf("Oops! %s\n", err)
		os.Exit(1)
	}

	next := cxt.Get("subcommand", "help").(string)
	if router.HasRoute(next) {
		if err := router.HandleRequest(next, cxt, false); err != nil {
			fmt.Printf("Oops! %s\n", err)
			os.Exit(1)
		}
	} else {
		if err := router.HandleRequest("@plugin", cxt, false); err != nil {
			fmt.Printf("Oops! %s\n", err)
			os.Exit(1)
		}
	}

}

func routes(reg *cookoo.Registry, cxt cookoo.Context) {

	flags := flag.NewFlagSet("global", flag.PanicOnError)
	flags.Bool("h", false, "Print help text.")

	cxt.Put("os.Args", os.Args)

	reg.Route("@startup", "Parse args and send to the right subcommand.").
		Does(cli.ShiftArgs, "_").Using("n").WithDefault(1).
		Does(cli.ParseArgs, "parseargs").
		Using("flagset").WithDefault(flags).
		Using("args").From("cxt:os.Args").
		Does(cli.ShowHelp, "help").
		Using("show").From("cxt:h cxt:help").
		Using("summary").WithDefault(Summary).
		Using("usage").WithDefault(Usage).
		Using("flags").WithDefault(flags).
		Does(parseConfig, "cfg").
		Does(subcommand, "subcommand").
		Using("args").From("cxt:os.Args")

	reg.Route("help", "Print help.").
		Does(cli.ShowHelp, "help").
		Using("show").WithDefault(true).
		Using("summary").WithDefault(Summary).
		Using("usage").WithDefault(Usage).
		Using("flags").WithDefault(flags)

	reg.Route("in", "Set GOPATH and supporting env vars.").Does(configEnv, "gopath")
	reg.Route("out", "Set GOPATH back to former val.").Does(out, "gopath")

	reg.Route("install", "Install dependencies.").
	Does(mkdir, "dir").Using("dir").WithDefault("_vendor").
	Does(linkPkg, "alias").
	Does(deps, "dependencies").Using("cfg").From("cxt:cfg")

	reg.Route("@plugin", "Try to send to a plugin.").
		Does(func (c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
			fmt.Printf("Command '%s' not found.", c.Get("subcommand", "").(string))
			return nil, nil
		}, "_")
}

func subcommand(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	args := p.Get("args", []string{"help"}).([]string)
	if len(args) == 0 {
		return "help", nil
	}
	return args[0], nil
}

func parseConfig(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	f, err := yaml.ReadFile("./glide.yaml")
	if err != nil {
		return nil, err
	}

	// Convenience:
	top, ok := f.Root.(yaml.Map)
	if !ok {
		return nil, fmt.Errorf("Expected YAML root to be map, got %t", f.Root)
	}

	vals := map[string]yaml.Node(top)
	if name, ok := vals["package"]; ok {
		c.Put("cfg.package", name.(yaml.Scalar).String())
	}
	if imp, ok := vals["import"]; ok {
		fmt.Printf("Imports: %v\n", imp)
		c.Put("cfg.import", imp)
	}

	return f, nil
}

func configEnv(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	gopath := fmt.Sprintf("%s/_vendor", cwd)
	binpath := fmt.Sprintf("$PATH:%s/bin", gopath)
	old_gopath := os.Getenv("GOPATH")
	old_binpath := os.Getenv("PATH")

	os.Setenv("OLD_GOPATH", old_gopath)
	os.Setenv("OLD_PATH", old_binpath)
	os.Setenv("GOPATH", gopath)
	os.Setenv("PATH_TEST", binpath)
	fmt.Printf("export GOPATH=%s\n", gopath)

	return nil, nil
}

func out(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	os.Setenv("GOPATH", os.Getenv("OLD_GOPATH"))
	os.Setenv("PATH_TEST", os.Getenv("OLD_PATH"))
	return true, nil
}

func deps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	//cfg := p.Get("cfg", nil).(*yaml.File)
	imports := c.Get("cfg.import", nil)
	if imports == nil {
		fmt.Printf("[INFO] No dependencies found. Nothing downloaded.")
		return false, nil
	}
	imp, ok := imports.(yaml.Map)
	if !ok {
		return nil, fmt.Errorf("Malformed YAML: Expected imports to be a map.")
	}

	for k, v := range imp {
		fmt.Printf("Key: %s, Value: %v\n", k, v)
		vmap, ok := v.(yaml.Map)
		if ok {
			settings := map[string]yaml.Node(vmap)
			if repo, ok := settings["repo"]; ok {
				fmt.Printf("Not implemented: Fetch repo %s\n", repo)
			} else {
				exec.Command("go", "get", k)
			}
		} else {
			simpleName := string(k)
			fmt.Printf("[INFO] Installing %s\n", simpleName)
			if err := exec.Command("go", "get", simpleName).Run(); err != nil {
				fmt.Printf("[WARN] Failed to install %s: %s", simpleName, err)
			}
		}
	}

	return nil, nil
}

func mkdir(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	target := os.Getenv("GOPATH")
	//pname := c.Get("cfg.package", "").(string)
	if len(target) == 0 {
		return nil, fmt.Errorf("$GOPATH appears to be unset.")
	}

	target = fmt.Sprintf("%s/src", target)

	if err := os.MkdirAll(target, os.ModeDir | 0755); err != nil {
		return nil, fmt.Errorf("Failed to make directory %s: %s", target, err)
	}

	return nil, nil
}

func linkPkg(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	pname := c.Get("cfg.package", "").(string)

	here, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Could not get current directory: %s", err)
	}

	gopath := os.Getenv("GOPATH")
	if len(gopath) == 0 {
		return nil, fmt.Errorf("$GOPATH appears to be unset.")
	}
	if len(pname) == 0 {
		return nil, fmt.Errorf("glide.yaml is missing 'package:'")
	}

	base := path.Dir(pname)
	if base != "." {
		dir := fmt.Sprintf("%s/src/%s", gopath, base)
		if err := os.MkdirAll(dir, os.ModeDir | 0755); err != nil {
			return nil, fmt.Errorf("Failed to make directory %s: %s", dir, err)
		}
	}

	ldest := fmt.Sprintf("%s/src/%s", gopath, pname)
	if err := os.Symlink(here, ldest); err != nil {
		if os.IsExist(err) {
			fmt.Printf("[INFO] Link to %s already exists. Skipping.\n", ldest)
		} else {
			return nil, fmt.Errorf("Failed to create symlink from %s to %s: %s", gopath, ldest, err)
		}
	}

	return ldest, nil
}
