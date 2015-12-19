package action

import (
	"os"
	"path/filepath"
	"testing"
)

const testdata = "../testdata/path"

func TestGlideWD(t *testing.T) {
	wd := filepath.Join(testdata, "a/b/c")
	found, err := GlideWD(wd)
	if err != nil {
		t.Errorf("Failed to get Glide directory: %s", err)
	}

	if found != filepath.Join(testdata, "a") {
		t.Errorf("Expected %s to match %s", found, filepath.Join(wd, "a"))
	}

	// This should fail
	wd = "/No/Such/Dir"
	found, err = GlideWD(wd)
	if err == nil {
		t.Errorf("Expected to get an error on a non-existent directory, not %s", found)
	}

}

func TestVendorPath(t *testing.T) {
	td, err := filepath.Abs(testdata)
	if err != nil {
		t.Fatal(err)
	}
	os.Chdir(filepath.Join(td, "a/b/c"))
	res, err := VendorPath()
	if err != nil {
		t.Errorf("Failed to resolve vendor directory: %s", err)
	}
	expect := filepath.Join(td, "a", "vendor")
	if res != expect {
		t.Errorf("Failed to find vendor: expected %s got %s", expect, res)
	}
}
