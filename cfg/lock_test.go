package cfg

import (
	"sort"
	"testing"
)

func TestSortLocks(t *testing.T) {
	c, err := ConfigFromYaml([]byte(yml))
	if err != nil {
		t.Error("ConfigFromYaml failed to parse yaml for TestSortDependencies")
	}

	ls := make(Locks, len(c.Imports))
	for i := 0; i < len(c.Imports); i++ {
		ls[i] = &Lock{
			Name:    c.Imports[i].Name,
			Version: c.Imports[i].Reference,
		}
	}

	if ls[2].Name != "github.com/Masterminds/structable" {
		t.Error("Initial dependencies are out of order prior to sort")
	}

	sort.Sort(ls)

	if ls[0].Name != "github.com/kylelemons/go-gypsy" ||
		ls[1].Name != "github.com/Masterminds/convert" ||
		ls[2].Name != "github.com/Masterminds/cookoo" ||
		ls[3].Name != "github.com/Masterminds/structable" {
		t.Error("Sorting of dependencies failed")
	}
}
