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

// Filter a list of versions to only included semantic versions. The response
// is a mapping of the original version to the semantic version.
func getSemVers(refs []string) map[string]string {
	sv := map[string]string{}
	for _, r := range refs {
		nv, err := filterVersion(r)
		if err == nil {
			sv[r] = nv
		}
	}

	return sv
}

// Get all the references for a repo. This includes the tags and branches.
func getAllVcsRefs(repo vcs.Repo) ([]string, error) {
	tags, err := repo.Tags()
	if err != nil {
		return []string{}, err
	}

	branches, err := repo.Branches()
	if err != nil {
		return []string{}, err
	}

	return append(branches, tags...), nil
}

func isBranch(branch string, repo vcs.Repo) (bool, error) {
	branches, err := repo.Branches()
	if err != nil {
		return false, err
	}
	for _, b := range branches {
		if b == branch {
			return true, nil
		}
	}
	return false, nil
}
