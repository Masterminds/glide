package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Masterminds/cookoo"
)

// This file contains commands for working with Godep.

// The Godeps struct from Godep.
//
// https://raw.githubusercontent.com/tools/godep/master/dep.go
//
// We had to copy this because it's in the package main for Godep.
type Godeps struct {
	ImportPath string
	GoVersion  string
	Packages   []string `json:",omitempty"` // Arguments to save, if any.
	Deps       []GodepDependency

	outerRoot string
}

// GodepDependency is a modified version of Godep's Dependency struct.
// It drops all of the unexported fields.
type GodepDependency struct {
	ImportPath string
	Comment    string `json:",omitempty"` // Description of commit, if present.
	Rev        string // VCS-specific commit ID.
}

func HasGodepGodeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	dir := cookoo.GetString("dir", "", p)
	path := filepath.Join(dir, "Godeps/Godeps.json")
	_, err := os.Stat(path)
	return err == nil, nil
}

// Parse the Godep Godeps.json file.
//
// Params:
// - dir (string): the project's directory
//
// Returns an []*Dependency
func ParseGodepGodeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	dir := cookoo.GetString("dir", "", p)
	path := filepath.Join(dir, "Godeps/Godeps.json")
	if _, err := os.Stat(path); err != nil {
		return []*Dependency{}, nil
	}
	// Info("Found Godeps.json file.\n")

	buf := []*Dependency{}

	godeps := new(Godeps)

	// Get a handle to the file.
	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	defer file.Close()


	dec := json.NewDecoder(file)
	if err := dec.Decode(godeps); err != nil {
		return buf, err
	}

	// Info("Importing %d packages from %s.\n", len(godeps.Deps), godeps.ImportPath)

	for _, d := range godeps.Deps {
		// Info("Adding package %s\n", d.ImportPath)
		dep := &Dependency{ Name: d.ImportPath, Reference: d.Rev }
		buf = append(buf, dep)
	}

	return buf, nil
}
