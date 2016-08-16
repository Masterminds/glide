package mirrors

import (
	"io/ioutil"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Mirrors contains global mirrors to local configuration
type Mirrors struct {

	// Repos contains repo mirror configuration
	Repos MirrorRepos `yaml:"repos"`
}

// Marshal converts a Mirror instance to YAML
func (ov *Mirrors) Marshal() ([]byte, error) {
	yml, err := yaml.Marshal(&ov)
	if err != nil {
		return []byte{}, err
	}
	return yml, nil
}

// WriteFile writes an mirrors.yaml file
//
// This is a convenience function that marshals the YAML and then writes it to
// the given file. If the file exists, it will be clobbered.
func (ov *Mirrors) WriteFile(opath string) error {
	o, err := ov.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(opath, o, 0666)
}

// ReadMirrorsFile loads the contents of an mirrors.yaml file.
func ReadMirrorsFile(opath string) (*Mirrors, error) {
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

// FromYaml returns an instance of Mirrors from YAML
func FromYaml(yml []byte) (*Mirrors, error) {
	ov := &Mirrors{}
	err := yaml.Unmarshal([]byte(yml), &ov)
	return ov, err
}

// MarshalYAML is a hook for gopkg.in/yaml.v2.
// It sorts mirror repos lexicographically for reproducibility.
func (ov *Mirrors) MarshalYAML() (interface{}, error) {

	sort.Sort(ov.Repos)

	return ov, nil
}

// MirrorRepos is a slice of Mirror pointers
type MirrorRepos []*MirrorRepo

// Len returns the length of the MirrorRepos. This is needed for sorting with
// the sort package.
func (o MirrorRepos) Len() int {
	return len(o)
}

// Less is needed for the sort interface. It compares two MirrorRepos based on
// their original value.
func (o MirrorRepos) Less(i, j int) bool {

	// Names are normalized to lowercase because case affects sorting order. For
	// example, Masterminds comes before kylelemons. Making them lowercase
	// causes kylelemons to come first which is what is expected.
	return strings.ToLower(o[i].Original) < strings.ToLower(o[j].Original)
}

// Swap is needed for the sort interface. It swaps the position of two
// MirrorRepos.
func (o MirrorRepos) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

// MirrorRepo represents a single repo mirror
type MirrorRepo struct {
	Original string `yaml:"original"`
	Repo     string `yaml:"repo"`
	Vcs      string `yaml:"vcs,omitempty"`
}
