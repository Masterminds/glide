package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
	"os"
	//"os/user"
	//"os/exec"
)

// AlreadyGliding emits a warning (and stops) if we're in a glide session.
//
// This should be used when you want to make sure that we're not already in a
// glide environment.
func AlreadyGliding(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	if os.Getenv("ALREADY_GLIDING") == "1" {
		fmt.Printf("[WARN] You're already gliding. Run `glide out` to stop your current glide.\n")
		return true, &cookoo.Stop{}
	}
	return false, nil
}

// ReadyToGlide fails if the environment is not sufficient for using glide.
//
// Most importantly, it fails if glide.yaml is not present in the current
// working directory.
func ReadyToGlide(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	if _, err := os.Stat("./glide.yaml"); err != nil {
		cwd, _ := os.Getwd()
		return false, fmt.Errorf("glide.yaml is missing from %s", cwd)
	}
	return true, nil
}

func In(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	gopath := fmt.Sprintf("%s/_vendor", cwd)

	fmt.Printf("export OLD_PATH=%s\n", os.Getenv("PATH"))
	fmt.Printf("export PATH=%s:%s\n", os.Getenv("PATH"), gopath + "/bin")
	fmt.Printf("export GOPATH=%s\n", gopath)
	fmt.Printf("export ALREADY_GLIDING=1\n")

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

	into := p.Get("into", "").(string)
	fmt.Printf("Gliding into %s\n", into)
	if len(into) > 0 {
		if err := os.Chdir(into); err != nil {
			return nil, err
		}
	}

	shell := os.Getenv("SHELL")
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
	gopath := fmt.Sprintf("%s/_vendor", cwd)

	os.Setenv("ALREADY_GLIDING", "1")
	os.Setenv("GOPATH", gopath)
	os.Setenv("PATH", path + ":" + gopath + "/bin")

	pa := os.ProcAttr {
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir: cwd,
	}

	/*
	loginPath, err := exec.LookPath("login")
	if err != nil {
		return nil, err
	}
	*/

	fmt.Printf(">> You are now gliding into a new shell. To exit, type 'exit'\n")
	// Login may work better than executing the shell manually.
	//proc, err := os.StartProcess(loginPath, []string{"login", "-fpl", u.Username}, &pa)
	proc, err := os.StartProcess(shell, []string{shell}, &pa)
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

func Out(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	/*
	os.Setenv("GOPATH", os.Getenv("OLD_GOPATH"))
	os.Setenv("PATH_TEST", os.Getenv("OLD_PATH"))

	fmt.Printf("export GOPATH=%s\n", os.Getenv("OLD_GOPATH"))
	fmt.Printf("export PATH=%s\n", os.Getenv("OLD_PATH"))
	fmt.Printf("export OLD_GOPATH=\n")
	fmt.Printf("export OLD_PATH=\n")
	fmt.Printf("export ALREADY_GLIDING=\n")
	*/
	if os.Getenv("ALREADY_GLIDING") != "1" {
		fmt.Println("You are not currently gliding. To begin, try 'glide in'.")
		return false, nil
	}
	fmt.Printf("To exit this glide, type 'exit'\n")
	return true, nil
}
