package action

import (
	"io/ioutil"
	"os"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
)

// EnsureConfig loads and returns a config file.
//
// Any error will cause an immediate exit, with an error printed to Stderr.
func EnsureConfig(yamlpath string) *cfg.Config {
	yml, err := ioutil.ReadFile(yamlpath)
	if err != nil {
		msg.Error("Failed to load %s: %s", yamlpath, err)
		os.Exit(2)
	}
	conf, err := cfg.ConfigFromYaml(yml)
	if err != nil {
		msg.Error("Failed to parse %s: %s", yamlpath, err)
		os.Exit(3)
	}

	return conf
}

func EnsureCacheDir() {
	msg.Warn("ensure.go: ensureCacheDir is not implemented.")
}
