package action

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/mirrors"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

// MirrorsList displays a list of currently setup mirrors.
func MirrorsList() error {
	home := gpath.Home()

	op := filepath.Join(home, "mirrors.yaml")

	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Info("No mirrors exist. No mirrors.yaml file not found")
		return nil
	}

	ov, err := mirrors.ReadMirrorsFile(op)
	if err != nil {
		msg.Die("Unable to read mirrors.yaml file: %s", err)
	}

	if len(ov.Repos) == 0 {
		msg.Info("No mirrors found")
		return nil
	}

	msg.Info("Mirrors...")
	for _, r := range ov.Repos {
		if r.Vcs == "" {
			msg.Info("--> %s replaced by %s", r.Original, r.Repo)
		} else {
			msg.Info("--> %s replaced by %s (%s)", r.Original, r.Repo, r.Vcs)
		}
	}

	return nil
}

// MirrorsSet sets a mirror to use
func MirrorsSet(o, r, v string) error {
	if o == "" || r == "" {
		msg.Err("Both the original and mirror values are required")
		return nil
	}

	home := gpath.Home()

	op := filepath.Join(home, "mirrors.yaml")

	var ov *mirrors.Mirrors
	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Info("No mirrors.yaml file exists. Creating new one")
		ov = &mirrors.Mirrors{
			Repos: make(mirrors.MirrorRepos, 0),
		}
	} else {
		ov, err = mirrors.ReadMirrorsFile(op)
		if err != nil {
			msg.Die("Error reading existing mirrors.yaml file: %s", err)
		}
	}

	found := false
	for i, re := range ov.Repos {
		if re.Original == o {
			found = true
			msg.Info("%s found in mirrors. Replacing with new settings", o)
			ov.Repos[i].Repo = r
			ov.Repos[i].Vcs = v
		}
	}

	if !found {
		nr := &mirrors.MirrorRepo{
			Original: o,
			Repo:     r,
			Vcs:      v,
		}
		ov.Repos = append(ov.Repos, nr)
	}

	msg.Info("%s being set to %s", o, r)

	err := ov.WriteFile(op)
	if err != nil {
		msg.Err("Error writing mirrors.yaml file: %s", err)
	} else {
		msg.Info("mirrors.yaml written with changes")
	}

	return nil
}

// MirrorsRemove removes a mirrors setting
func MirrorsRemove(k string) error {
	if k == "" {
		msg.Err("The mirror to remove is required")
		return nil
	}

	home := gpath.Home()

	op := filepath.Join(home, "mirrors.yaml")

	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Err("mirrors.yaml file not found")
		return nil
	}

	ov, err := mirrors.ReadMirrorsFile(op)
	if err != nil {
		msg.Die("Unable to read mirrors.yaml file: %s", err)
	}

	var nre mirrors.MirrorRepos
	var found bool
	for _, re := range ov.Repos {
		if re.Original != k {
			nre = append(nre, re)
		} else {
			found = true
		}
	}

	if !found {
		msg.Warn("%s was not found in mirrors", k)
	} else {
		msg.Info("%s was removed from mirrors", k)
		ov.Repos = nre

		err = ov.WriteFile(op)
		if err != nil {
			msg.Err("Error writing mirrors.yaml file: %s", err)
		} else {
			msg.Info("mirrors.yaml written with changes")
		}
	}

	return nil
}
