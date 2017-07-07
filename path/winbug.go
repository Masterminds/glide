package path

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"bytes"
	"io/ioutil"

	"github.com/Masterminds/glide/msg"
)

// This file and its contents are to handle a Windows bug where large sets of
// files fail when using the `os` package. This has been seen in Windows 10
// including the Windows Linux Subsystem.
// Tracking the issue in https://github.com/golang/go/issues/20841. Once the
// upstream issue is fixed this change can be reverted.

// CustomRemoveAll is similar to os.RemoveAll but deals with the bug outlined
// at https://github.com/golang/go/issues/20841.
func CustomRemoveAll(p string) error {

	// Handle the windows case first
	if runtime.GOOS == "windows" {
		msg.Debug("Detected Windows. Removing files using windows command")
		cmd := exec.Command("cmd.exe", "/c", fmt.Sprintf("rd /s /q %s", p))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Error removing files: %s. output: %s", err, output)
		}
	} else if detectWsl() {
		cmd := exec.Command("rm", "-rf", p)
		output, err2 := cmd.CombinedOutput()
		msg.Debug("Detected Windows Subsystem for Linux. Removing files using subsystem command")
		if err2 != nil {
			return fmt.Errorf("Error removing files: %s. output: %s", err2, output)
		}
	}
	return os.RemoveAll(p)
}

// CustomRename is similar to os.Rename but deals with the bug outlined
// at https://github.com/golang/go/issues/20841.
func CustomRename(o, n string) error {

	// Handking windows cases first
	if runtime.GOOS == "windows" {
		msg.Debug("Detected Windows. Moving files using windows command")
		cmd := exec.Command("cmd.exe", "/c", "move", o, n)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Error moving files: %s. output: %s", err, output)
		}

		return nil
	} else if detectWsl() {
		cmd := exec.Command("mv", o, n)
		output, err2 := cmd.CombinedOutput()
		msg.Debug("Detected Windows Subsystem for Linux. Removing files using subsystem command")
		if err2 != nil {
			return fmt.Errorf("Error moving files: %s. output: %s", err2, output)
		}

		return nil
	}

	return os.Rename(o, n)
}

var procIsWin bool
var procDet bool

func detectWsl() bool {

	if !procDet {
		procDet = true
		_, err := os.Stat("/proc/version")
		if err == nil {
			b, err := ioutil.ReadFile("/proc/version")
			if err != nil {
				msg.Warn("Unable to read /proc/version that was detected. May incorrectly detect WSL")
				msg.Debug("Windows Subsystem for Linux detection error: %s", err)
				return false
			}

			if bytes.Contains(b, []byte("Microsoft")) {
				msg.Debug("Windows Subsystem for Linux detected")
				procIsWin = true
			}
		}
	}

	return procIsWin
}
