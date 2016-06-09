// Package gpm reads GPM's Godeps files.
//
// It is not a complete implementaton of GPM.
package gpm

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

// Has indicates whether a Godeps file exists.
func Has(dir string) bool {
	path := filepath.Join(dir, "Godeps")
	_, err := os.Stat(path)
	return err == nil
}

// Parse parses a GPM-flavored Godeps file.
func Parse(dir string) ([]*cfg.Dependency, error) {
	path := filepath.Join(dir, "Godeps")
	if i, err := os.Stat(path); err != nil {
		return []*cfg.Dependency{}, nil
	} else if i.IsDir() {
		msg.Info("Godeps is a directory. This is probably a Godep project.\n")
		return []*cfg.Dependency{}, nil
	}
	msg.Info("Found Godeps file in %s", gpath.StripBasepath(dir))
	msg.Info("--> Parsing GPM metadata...")

	buf := []*cfg.Dependency{}

	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts, ok := parseGodepsLine(scanner.Text())
		if ok {
			dep := &cfg.Dependency{Name: parts[0]}
			if len(parts) > 1 {
				dep.Reference = parts[1]
			}
			buf = append(buf, dep)
		}
	}
	if err := scanner.Err(); err != nil {
		msg.Warn("Scan failed: %s\n", err)
		return buf, err
	}

	return buf, nil
}

func parseGodepsLine(line string) ([]string, bool) {
	line = strings.TrimSpace(line)

	if len(line) == 0 || strings.HasPrefix(line, "#") {
		return []string{}, false
	}

	return strings.Fields(line), true
}
