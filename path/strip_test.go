package path

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestStripVcs(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "strip-vcs")
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	// Make VCS directories.
	gp := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gp, 0755)
	if err != nil {
		t.Error(err)
	}

	bp := filepath.Join(tempDir, ".bzr")
	err = os.Mkdir(bp, 0755)
	if err != nil {
		t.Error(err)
	}

	hp := filepath.Join(tempDir, ".hg")
	err = os.Mkdir(hp, 0755)
	if err != nil {
		t.Error(err)
	}

	sp := filepath.Join(tempDir, ".svn")
	err = os.Mkdir(sp, 0755)
	if err != nil {
		t.Error(err)
	}

	ov := VendorDir
	VendorDir = tempDir

	StripVcs()

	VendorDir = ov

	if _, err := os.Stat(gp); !os.IsNotExist(err) {
		t.Error(".git directory not deleted")
	}
	if _, err := os.Stat(hp); !os.IsNotExist(err) {
		t.Error(".hg directory not deleted")
	}
	if _, err := os.Stat(bp); !os.IsNotExist(err) {
		t.Error(".bzr directory not deleted")
	}
	if _, err := os.Stat(sp); !os.IsNotExist(err) {
		t.Error(".svn directory not deleted")
	}
}
