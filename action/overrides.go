package action

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/msg"
	"github.com/Masterminds/glide/overrides"
	gpath "github.com/Masterminds/glide/path"
)

// OverridesList displays a list of currently setup overrides.
func OverridesList() error {
	home := gpath.Home()

	op := filepath.Join(home, "overrides.yaml")

	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Info("No overrides exist. No overrides.yaml file not found")
		return nil
	}

	ov, err := overrides.ReadOverridesFile(op)
	if err != nil {
		msg.Die("Unable to read overrides.yaml file: %s", err)
	}

	if len(ov.Repos) == 0 {
		msg.Info("No overrides found")
		return nil
	}

	msg.Info("Overrides...")
	for _, r := range ov.Repos {
		if r.Vcs == "" {
			msg.Info("--> %s replaced by %s", r.Original, r.Repo)
		} else {
			msg.Info("--> %s replaced by %s (%s)", r.Original, r.Repo, r.Vcs)
		}
	}

	return nil
}

// OverridesSet sets an override to use
func OverridesSet(o, r, v string) error {
	if o == "" || r == "" {
		msg.Err("Both the original and overriding values are required")
		return nil
	}

	home := gpath.Home()

	op := filepath.Join(home, "overrides.yaml")

	var ov *overrides.Overrides
	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Info("No overrides.yaml file exists. Creating new one")
		ov = &overrides.Overrides{
			Repos: make(overrides.OverrideRepos, 0),
		}
	} else {
		ov, err = overrides.ReadOverridesFile(op)
		if err != nil {
			msg.Die("Error reading existing overrides.yaml file: %s", err)
		}
	}

	found := false
	for i, re := range ov.Repos {
		if re.Original == o {
			found = true
			msg.Info("%s found in overrides. Replacing with new settings", o)
			ov.Repos[i].Repo = r
			ov.Repos[i].Vcs = v
		}
	}

	if !found {
		nr := &overrides.OverrideRepo{
			Original: o,
			Repo:     r,
			Vcs:      v,
		}
		ov.Repos = append(ov.Repos, nr)
	}

	msg.Info("%s being set to %s", o, r)

	err := ov.WriteFile(op)
	if err != nil {
		msg.Err("Error writing overrides.yaml file: %s", err)
	} else {
		msg.Info("overrides.yaml written with changes")
	}

	return nil
}

// OverridesRemove removes an override setting
func OverridesRemove(k string) error {
	if k == "" {
		msg.Err("The override to remove is required")
		return nil
	}

	home := gpath.Home()

	op := filepath.Join(home, "overrides.yaml")

	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Err("overrides.yaml file not found")
		return nil
	}

	ov, err := overrides.ReadOverridesFile(op)
	if err != nil {
		msg.Die("Unable to read overrides.yaml file: %s", err)
	}

	var nre overrides.OverrideRepos
	var found bool
	for _, re := range ov.Repos {
		if re.Original != k {
			nre = append(nre, re)
		} else {
			found = true
		}
	}

	if !found {
		msg.Warn("%s was not found in overrides", k)
	} else {
		msg.Info("%s was removed from overrides", k)
		ov.Repos = nre

		err = ov.WriteFile(op)
		if err != nil {
			msg.Err("Error writing overrides.yaml file: %s", err)
		} else {
			msg.Info("overrides.yaml written with changes")
		}
	}

	return nil
}
