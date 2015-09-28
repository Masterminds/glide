package cmd

import (
	"errors"
	"regexp"

	"github.com/Masterminds/vcs"
)

// The SemVer handling by github.com/hashicorp/go-version provides the ability
// to work with versions

// The compiled regular expression used to test the validity of a version.
var versionRegexp *regexp.Regexp

const versionRegexpRaw string = `v?(([0-9]+(\.[0-9]+){0,2})` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?)` +
	`?`

func init() {
	versionRegexp = regexp.MustCompile("^" + versionRegexpRaw + "$")
}

// Filter the leading v from the version. Returns an error if there
// was an issue including if the version was not SemVer
// returns:
// - the semantic verions (stripping any leading v if present)
// - error if there was one
func filterVersion(v string) (string, error) {
	matches := versionRegexp.FindStringSubmatch(v)
	if matches == nil || matches[1] == "" {
		return "", errors.New("No SemVer found.")
	}
	return matches[1], nil
}

// Get all the references for a repo. This includes the tags and branches.
func getAllVcsRefs(repo vcs.Repo) ([]string, error) {
	refs := []string{}

	tags, err := repo.Tags()
	if err != nil {
		return []string{}, err
	}
	for _, ref := range tags {
		refs = append(refs, ref)
	}

	branches, err := repo.Branches()
	if err != nil {
		return []string{}, err
	}
	for _, ref := range branches {
		refs = append(refs, ref)
	}

	return refs, nil
}

// From the refs find all of the ones fitting the SemVer pattern.
// func findSemVerRefs(refs []string) ([]string, error) {
//
// }
