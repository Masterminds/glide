package cmd

import (
	"github.com/Masterminds/cookoo"
)

// UpdateRevisions updates the revision numbers on all of the imports.
func UpdateReferences(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", &Config{}).(*Config)

	if len(cfg.Imports) == 0 {
		return cfg, nil
	}

	for _, imp := range cfg.Imports {
		commit, err := VcsLastCommit(imp)
		if err != nil {
			Warn("Could not get commit on %s: %s", imp.Name, err)
		}
		imp.Reference = commit
	}

	return cfg, nil
}
