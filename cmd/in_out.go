package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
	"os"
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

func Out(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	os.Setenv("GOPATH", os.Getenv("OLD_GOPATH"))
	os.Setenv("PATH_TEST", os.Getenv("OLD_PATH"))

	fmt.Printf("export GOPATH=%s\n", os.Getenv("OLD_GOPATH"))
	fmt.Printf("export PATH=%s\n", os.Getenv("OLD_PATH"))
	fmt.Printf("export OLD_GOPATH=\n")
	fmt.Printf("export OLD_PATH=\n")
	fmt.Printf("export ALREADY_GLIDING=\n")
	return true, nil
}
