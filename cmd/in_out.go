package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/cookoo"
)

// AlreadyGliding emits a warning (and stops) if we're in a glide session.
//
// This should be used when you want to make sure that we're not already in a
// glide environment.
func AlreadyGliding(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	if os.Getenv("ALREADY_GLIDING") == "1" {
		Warn("You're already gliding. Use `exit` to leave the glide shell.\n")
		return true, &cookoo.Stop{}
	}
	return false, nil
}

// ReadyToGlide fails if the environment is not sufficient for using glide.
//
// Most importantly, it fails if glide.yaml is not present in the current
// working directory.
func ReadyToGlide(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)
	if _, err := os.Stat(fname); err != nil {
		cwd, _ := os.Getwd()
		return false, fmt.Errorf("%s is missing from %s", fname, cwd)
	}
	return true, nil
}

// GlideGopath returns the GOPATH for a Glide project.
//
// It determines the GOPATH by searching for the glide.yaml file, and then
// assuming the vendor/ directory is in that directory. It traverses
// the tree upwards (e.g. only ancestors).
//
// If no glide.yaml is found, or if a directory cannot be read, this returns
// an error.
func GlideGopath(c cookoo.Context, filename string) (string, error) {
	vendor := c.Get("VendorDir", "vendor").(string)
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Find the directory that contains glide.yaml
	yamldir, err := glideWD(cwd, filename)
	if err != nil {
		return cwd, err
	}

	//gopath := fmt.Sprintf("%s/_vendor", yamldir)
	gopath := filepath.Join(yamldir, vendor)

	return gopath, nil
}

// Return the path to the vendor directory.
func VendorPath(c cookoo.Context) (string, error) {
	vendor := c.Get("VendorDir", "vendor").(string)
	filename := c.Get("yaml", "glide.yaml").(string)
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Find the directory that contains glide.yaml
	yamldir, err := glideWD(cwd, filename)
	if err != nil {
		return cwd, err
	}

	gopath := filepath.Join(yamldir, vendor)

	return gopath, nil
}

func glideWD(dir, filename string) (string, error) {
	fullpath := filepath.Join(dir, filename)

	if _, err := os.Stat(fullpath); err == nil {
		return dir, nil
	}

	base := filepath.Dir(dir)
	if base == dir {
		return "", fmt.Errorf("Cannot resolve parent of %s", base)
	}

	return glideWD(base, filename)
}

// In emits GOPATH for editors and such.
func In(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)

	gopath, err := GlideGopath(c, fname)
	if err != nil {
		return nil, err
	}

	/*
		fmt.Printf("export OLD_PATH=%s\n", os.Getenv("PATH"))
		fmt.Printf("export PATH=%s:%s\n", gopath + "/bin", os.Getenv("PATH"))
		fmt.Printf("export GOPATH=%s\n", gopath)
		fmt.Printf("export ALREADY_GLIDING=1\n")
	*/
	fmt.Println(gopath)

	return nil, nil
}

// Into starts a new shell as a child of glide.
// This new shell inherits the environment typical of a Glide In, but
// without any shell export weirdness. Optionally, if a path is provided, this
// will glide into *that* directory.
//
// PARAMS
// 	- into (string): The directory to glide into.
func Into(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {

	cfg := p.Get("conf", &Config{}).(*Config)
	vendor := c.Get("VendorDir", "vendor").(string)

	into := p.Get("into", "").(string)
	if len(into) > 0 {
		if err := os.Chdir(into); err != nil {
			return nil, err
		}
	}

	// Shell and command args can be overwritten by config.InCommand.
	shell := os.Getenv("SHELL")
	cmdArgs := []string{shell}
	path := os.Getenv("PATH")
	/*
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
	*/

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	//gopath := fmt.Sprintf("%s/_vendor", cwd)
	gopath := filepath.Join(cwd, vendor)

	os.Setenv("ALREADY_GLIDING", "1")
	os.Setenv("GOPATH", gopath)
	os.Setenv("GOBIN", gopath+"/bin")
	os.Setenv("GLIDE_GOPATH", gopath)
	os.Setenv("PATH", gopath+"/bin:"+path)
	os.Setenv("GLIDE_PROJECT", cwd)
	os.Setenv("GLIDE_YAML", fmt.Sprintf("%s/glide.yaml", cwd))

	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
	}

	/*
		loginPath, err := exec.LookPath("login")
		if err != nil {
			return nil, err
		}
	*/

	// Allow incmd to override the Glide In default command.
	if len(cfg.InCommand) > 0 {
		cmdArgs = strings.Split(cfg.InCommand, " ")
		fmt.Printf(">> Running custom 'glide in': %v\n", cmdArgs)
		//shell, err = exec.LookPath(cmdArgs[0])
		//if err != nil {
		//return nil, err
		//}
		shell = cmdArgs[0]
	} else {
		fmt.Printf(">> You are now gliding into a new shell. To exit, type 'exit'\n")
	}

	if !filepath.IsAbs(shell) {
		shell, err = exec.LookPath(shell)
		if err != nil {
			return nil, err
		}
	}

	// Login may work better than executing the shell manually.
	//proc, err := os.StartProcess(loginPath, []string{"login", "-fpl", u.Username}, &pa)
	proc, err := os.StartProcess(shell, cmdArgs, &pa)
	if err != nil {
		return nil, err
	}

	state, err := proc.Wait()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Exited glide shell: %s", state.String())
	return nil, nil
}
