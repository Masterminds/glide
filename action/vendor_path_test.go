package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlideWD(t *testing.T) {
	cwd, _ := os.Getwd()
	filename := "glide.yaml"
	found, err := glideWD(cwd, filename)
	if err != nil {
		t.Errorf("Failed to get Glide directory: %s", err)
	}

	if found != filepath.Dir(cwd) {
		t.Errorf("Expected %s to match %s", found, filepath.Base(cwd))
	}

	// This should fail
	cwd = "/No/Such/Dir"
	found, err = glideWD(cwd, filename)
	if err == nil {
		t.Errorf("Expected to get an error on a non-existent directory, not %s", found)
	}

}
