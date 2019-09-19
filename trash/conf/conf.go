// Package govendor provides compatibility with govendor vendorfiles.

// This is a copy of Trash's `conf/conf.go` file, and has been ported
// to use gopkg.in/yaml.v2
// Trash is governed by a MIT license that can be found in the
// LICENSE file of Trash project

package conf

import (
	"bufio"
	"os"
	"sort"
	"strings"

	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Conf of a Trash vendor file.
type Conf struct {
	Package   string   `yaml:"package,omitempty"`
	Imports   []Import `yaml:"import,omitempty"`
	Excludes  []string `yaml:"exclude,omitempty"`
	importMap map[string]Import
	confFile  string
	yamlType  bool
}

// Import as defined in a Trash vendor file.
type Import struct {
	Package string `yaml:"package,omitempty"`
	Version string `yaml:"version,omitempty"`
	Repo    string `yaml:"repo,omitempty"`
}

// Parse a Trash vendor conf file.
func Parse(path string) (*Conf, error) {
	ymlfile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	trashConf := &Conf{confFile: path}
	err = yaml.Unmarshal(ymlfile, &trashConf)
	if err == nil {
		trashConf.yamlType = true
		trashConf.Dedupe()
		return trashConf, nil
	}

	trashConf = &Conf{confFile: path}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(bufio.NewReader(file))
	for scanner.Scan() {
		line := scanner.Text()
		if commentStart := strings.Index(line, "#"); commentStart >= 0 {
			line = line[0:commentStart]
		}
		if line = strings.TrimSpace(line); line == "" {
			continue
		}
		fields := strings.Fields(line)

		if len(fields) == 1 && trashConf.Package == "" {
			trashConf.Package = fields[0] // use the first 1-field line as the root package
			continue
		}
		// If we have a `-` suffix, it's an exclude pattern
		if fields[0][0] == '-' {
			trashConf.Excludes = append(trashConf.Excludes, strings.TrimSpace(fields[0][1:]))
			continue
		}
		// Otherwise it's an import pattern
		packageImport := Import{}
		packageImport.Package = fields[0] // at least 1 field at this point: trimmed the line and skipped empty
		if len(fields) > 2 {
			packageImport.Repo = fields[2]
		}
		if len(fields) > 1 {
			packageImport.Version = fields[1]
		}
		trashConf.Imports = append(trashConf.Imports, packageImport)
	}

	trashConf.Dedupe()
	return trashConf, nil
}

// Dedupe deletes duplicates and sorts the imports
func (t *Conf) Dedupe() {
	t.importMap = map[string]Import{}
	for _, i := range t.Imports {
		if _, ok := t.importMap[i.Package]; ok {
			continue
		}
		t.importMap[i.Package] = i
	}
	ps := make([]string, 0, len(t.importMap))
	for p := range t.importMap {
		ps = append(ps, p)
	}
	sort.Strings(ps)
	imports := make([]Import, 0, len(t.importMap))
	for _, p := range ps {
		imports = append(imports, t.importMap[p])
	}
	t.Imports = imports
}

// Get the import of a specified package.
func (t *Conf) Get(pkg string) (Import, bool) {
	i, ok := t.importMap[pkg]
	return i, ok
}

// ConfFile from which the config has been read.
func (t *Conf) ConfFile() string {
	return t.confFile
}
