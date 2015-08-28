package cmd

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestisDirectoryEmpty(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "empty-dir-test")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	empty, err := isDirectoryEmpty(tempDir)
	if err != nil {
		t.Error(err)
	}
	if empty == false {
		t.Error("isDirectoryEmpty reporting false on empty directory")
	}

	data := "foo bar baz"
	err = ioutil.WriteFile(tempDir+"/foo", []byte(data), 0644)
	if err != nil {
		t.Error(err)
	}

	empty, err = isDirectoryEmpty(tempDir)
	if err != nil {
		t.Error(err)
	}
	if empty == true {
		t.Error("isDirectoryEmpty reporting true on non-empty directory")
	}
}
