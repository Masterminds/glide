package cmd

import (
	"github.com/Masterminds/cookoo"
	"path/filepath"
	"bufio"
	"strings"
	"os"
)

// Indicates whether a Godeps file exists.
func HasGodeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	dir := cookoo.GetString("dir", "", p)
	path := filepath.Join(dir, "Godeps")
	_, err := os.Stat(path)
	return err == nil, nil
}


// Godeps parses a Godeps file.
//
// Params
// 	- dir (string): Directory root.
//
// Returns an []*Dependency
func Godeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	dir := cookoo.GetString("dir", "", p)
	path := filepath.Join(dir, "Godeps")
	if _, err := os.Stat(path); err != nil {
		return []*Dependency{}, nil
	}
	Info("Found Godeps file.\n")

	buf :=[]*Dependency{}

	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts, ok := parseGodepsLine(scanner.Text())
		if ok {
			dep := &Dependency{ Name: parts[0] }
			if len(parts) > 1 {
				dep.Reference = parts[1]
			}
			buf = append(buf, dep)
		}
	}
	if err := scanner.Err(); err != nil {
		Warn("Scan failed: %s\n", err)
		return buf, err
	}

	return buf, nil
}

func GodepsGit(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	dir := cookoo.GetString("dir", "", p)
	path := filepath.Join(dir, "Godeps-Git")
	if _, err := os.Stat(path); err != nil {
		return []*Dependency{}, nil
	}
	Info("Found Godeps-Git file.\n")

	buf :=[]*Dependency{}

	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts, ok := parseGodepsLine(scanner.Text())
		if ok {
			dep := &Dependency{ Name: parts[1], Repository: parts[0] }
			if len(parts) > 2 {
				dep.Reference = parts[2]
			}
			buf = append(buf, dep)
		}
	}
	if err := scanner.Err(); err != nil {
		Warn("Scan failed: %s\n", err)
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
