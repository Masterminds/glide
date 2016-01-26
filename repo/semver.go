package repo

import (
	"github.com/Masterminds/semver"
	"github.com/Masterminds/vcs"
)

// Filter a list of versions to only included semantic versions. The response
// is a mapping of the original version to the semantic version.
func getSemVers(refs []string) []*semver.Version {
	sv := []*semver.Version{}
	for _, r := range refs {
		v, err := semver.NewVersion(r)
		if err == nil {
			sv = append(sv, v)
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
