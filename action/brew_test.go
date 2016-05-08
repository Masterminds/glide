package action

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestBrew(t *testing.T) {
	// Capture stdout, revert when done
	var buf bytes.Buffer
	originalStdout := msg.Default.Stdout
	msg.Default.PanicOnDie = true
	msg.Default.Stdout = &buf
	defer func() {
		msg.Default.Stdout = originalStdout
	}()

	// Change to testdata dir for the duration of the test, and return when done
	originalDir, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get current directory: %s", err)
	}
	if err := os.Chdir("../testdata/brew"); err != nil {
		t.Errorf("Failed to change directory: %s", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory (%s): %s", originalDir, err)
		}
	}()

	Brew()

	// There should be exactly two resource blocks
	if strings.Count(buf.String(), "go_resource") != 2 {
		t.Error("Brew conversion created wrong number of resources")
	}

	// Resources should be named after the package path that will be vendored
	if !strings.Contains(buf.String(), `go_resource "github.com/Masterminds/semver"`) || !strings.Contains(buf.String(), `go_resource "a/different/path"`) {
		t.Error("Failed to name resources correctly")
	}

	// But both use the same repo, so there should be two of those
	if strings.Count(buf.String(), "https://github.com/Masterminds/semver") != 2 {
		t.Error("Faileded to set repo in resource correctly")
	}
}
