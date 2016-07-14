// Package gpm reads GPM's Godeps files.
//
// It is not a complete implementaton of GPM.
package gpm

import (
	"bufio"
	"fmt"
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

func AsMetadataPair(dir string) ([]*cfg.Dependency, *cfg.Lockfile, error) {
	path := filepath.Join(dir, "Godeps")
	if i, err := os.Stat(path); err != nil {
		return nil, nil, err
	} else if i.IsDir() {
		return nil, nil, fmt.Errorf("Found a Godeps dir, rather than it being a file")
	}

	var m []*cfg.Dependency
	l := &cfg.Lockfile{}

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts, ok := parseGodepsLine(scanner.Text())
		if ok {
			// Place no actual constraint on the project; rely instead on
			// gps's 'preferred version' reasoning from deps' lock
			// files...if we have one at all.
			if len(parts) > 1 {
				l.Imports = append(l.Imports, &cfg.Lock{Name: parts[0], Version: parts[1]})
			}
			m = append(m, &cfg.Dependency{Name: parts[0], Reference: "*"})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return m, l, nil
}

func parseGodepsLine(line string) ([]string, bool) {
	line = strings.TrimSpace(line)

	if len(line) == 0 || strings.HasPrefix(line, "#") {
		return []string{}, false
	}

	return strings.Fields(line), true
}
