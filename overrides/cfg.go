package overrides

import (
	"io/ioutil"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Overrides contains global overrides to local configuration
type Overrides struct {

	// Repos contains repo override configuration
	Repos OverrideRepos `yaml:"repos"`
}

// Marshal converts an Overrides instance to YAML
func (ov *Overrides) Marshal() ([]byte, error) {
	yml, err := yaml.Marshal(&ov)
	if err != nil {
		return []byte{}, err
	}
	return yml, nil
}

// WriteFile writes an overrides.yaml file
//
// This is a convenience function that marshals the YAML and then writes it to
// the given file. If the file exists, it will be clobbered.
func (ov *Overrides) WriteFile(opath string) error {
	o, err := ov.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(opath, o, 0666)
}

// ReadOverridesFile loads the contents of an overrides.yaml file.
func ReadOverridesFile(opath string) (*Overrides, error) {
	yml, err := ioutil.ReadFile(opath)
	if err != nil {
		return nil, err
	}
	ov, err := FromYaml(yml)
	if err != nil {
		return nil, err
	}
	return ov, nil
}

// FromYaml returns an instance of Overrides from YAML
func FromYaml(yml []byte) (*Overrides, error) {
	ov := &Overrides{}
	err := yaml.Unmarshal([]byte(yml), &ov)
	return ov, err
}

// MarshalYAML is a hook for gopkg.in/yaml.v2.
// It sorts override repos lexicographically for reproducibility.
func (ov *Overrides) MarshalYAML() (interface{}, error) {

	sort.Sort(ov.Repos)

	return ov, nil
}

// OverrideRepos is a slice of Override pointers
type OverrideRepos []*OverrideRepo

// Len returns the length of the OverrideRepos. This is needed for sorting with
// the sort package.
func (o OverrideRepos) Len() int {
	return len(o)
}

// Less is needed for the sort interface. It compares two OverrideRepos based on
// their original value.
func (o OverrideRepos) Less(i, j int) bool {

	// Names are normalized to lowercase because case affects sorting order. For
	// example, Masterminds comes before kylelemons. Making them lowercase
	// causes kylelemons to come first which is what is expected.
	return strings.ToLower(o[i].Original) < strings.ToLower(o[j].Original)
}

// Swap is needed for the sort interface. It swaps the position of two
// OverrideRepos.
func (o OverrideRepos) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

// OverrideRepo represents a single repo override
type OverrideRepo struct {
	Original string `yaml:"original"`
	Repo     string `yaml:"repo"`
	Vcs      string `yaml:"vcs,omitempty"`
}
