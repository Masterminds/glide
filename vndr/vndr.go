// Package vndr provides basic importing of Vendor.conf dependencies.
package vndr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// Has is a command to detect if a package contains a vendor.conf file.
func Has(dir string) bool {
	path := filepath.Join(dir, "vendor.conf")
	_, err := os.Stat(path)
	return err == nil
}

// Parse parses a VNDR vendor.conf file.
//
// It returns the contents as a dependency array.
func Parse(dir string) ([]*cfg.Dependency, error) {
	path := filepath.Join(dir, "vendor.conf")
	msg.Info(dir)
	if _, err := os.Stat(path); err != nil {
		return []*cfg.Dependency{}, nil
	}
	msg.Info("Found vendor.conf file in %s", gpath.StripBasepath(dir))
	msg.Info("--> Parsing vndr metadata...")

	buf := []*cfg.Dependency{}

	// Get a handle to the file.
	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	for s.Scan() {
		// Read and build the dependencies from the file
		ln := strings.TrimSpace(s.Text())
		if strings.HasPrefix(ln, "#") || ln == "" {
			continue
		}
		cidx := strings.Index(ln, "#")
		if cidx > 0 {
			ln = ln[:cidx]
		}
		ln = strings.TrimSpace(ln)
		parts := strings.Fields(ln)
		if len(parts) != 2 && len(parts) != 3 {
			return []*cfg.Dependency{}, fmt.Errorf("invalid config format: %s", ln)
		}

		dep := &cfg.Dependency{
			Reference: parts[1],
		}
		if len(parts) == 3 {
			dep.Repository = parts[2]
		}
		pkg, sub := util.NormalizeName(parts[0])
		dep.Name = pkg
		if len(sub) > 0 {
			dep.Subpackages = []string{sub}
		}

		buf = append(buf, dep)
	}
	if err := s.Err(); err != nil {
		return []*cfg.Dependency{}, err
	}

	return buf, nil
}
