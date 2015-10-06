package cmd

import "testing"

func TestFilterVersion(t *testing.T) {
	cases := []struct {
		version string
		semver  string
		err     bool
	}{
		{"1.2.3", "1.2.3", false},
		{"1.0", "1.0", false},
		{"1", "1", false},
		{"1.2.beta", "", true},
		{"foo", "", true},
		{"1.2-5", "1.2-5", false},
		{"1.2-beta.5", "1.2-beta.5", false},
		{"\n1.2", "", true},
		{"1.2.0-x.Y.0+metadata", "1.2.0-x.Y.0+metadata", false},
		{"1.2.0-x.Y.0+metadata-width-hypen", "1.2.0-x.Y.0+metadata-width-hypen", false},
		{"1.2.3-rc1-with-hypen", "1.2.3-rc1-with-hypen", false},
		{"1.2.3.4", "", true},
		{"v1.2.3", "1.2.3", false},
		{"foo1.2.3", "", true},
		{"v1.0", "1.0", false},
		{"v1", "1", false},
		{"v1.2.beta", "", true},
		{"v1.2-5", "1.2-5", false},
		{"v1.2-beta.5", "1.2-beta.5", false},
	}

	for _, tc := range cases {
		fv, err := filterVersion(tc.version)
		if tc.err && err == nil {
			t.Errorf("expected error for version: %s", tc.version)
		} else if !tc.err && err != nil {
			t.Errorf("error for version %s: %s", tc.version, err)
		}

		if tc.semver != fv {
			t.Errorf("expected version '%s' does not match actual version '%s'", tc.semver, fv)
		}
	}
}

func TestGetSemVers(t *testing.T) {
	versions := []string{
		"1.2.3",
		"1.0",
		"1",
		"1.2.beta",
		"foo",
		"1.2-5",
		"1.2-beta.5",
		"\n1.2",
		"1.2.0-x.Y.0+metadata",
		"1.2.0-x.Y.0+metadata-width-hypen",
		"1.2.3-rc1-with-hypen",
		"1.2.3.4",
		"v1.2.3",
		"foo1.2.3",
		"v1.0",
		"v1",
		"v1.2.beta",
		"v1.2-5",
		"v1.2-beta.5",
	}

	pass := map[string]string{
		"1.2.3":                            "1.2.3",
		"1.0":                              "1.0",
		"1":                                "1",
		"1.2-5":                            "1.2-5",
		"1.2-beta.5":                       "1.2-beta.5",
		"1.2.0-x.Y.0+metadata":             "1.2.0-x.Y.0+metadata",
		"1.2.0-x.Y.0+metadata-width-hypen": "1.2.0-x.Y.0+metadata-width-hypen",
		"1.2.3-rc1-with-hypen":             "1.2.3-rc1-with-hypen",
		"v1.2.3":                           "1.2.3",
		"v1.0":                             "1.0",
		"v1":                               "1",
		"v1.2-5":                           "1.2-5",
		"v1.2-beta.5":                      "1.2-beta.5",
	}

	sv := getSemVers(versions)
	for k, v := range sv {
		temp, ok := pass[v]
		if !ok {
			t.Errorf("GetSemVers found %s in error", k)
		}
		if k != temp {
			t.Errorf("GetSemVers found %s but expected %s", k, temp)
		}
	}
}

func TestGetSortedSemVerList(t *testing.T) {
	versions := []string{
		"1.2.3",
		"2.1",
		"2",
		"1.2-beta.5",
		"1.0",
		"2.0.3",
	}

	pass := []string{
		"2.1.0",
		"2.0.3",
		"2.0.0",
		"1.2.3",
		"1.2.0-beta.5",
		"1.0.0",
	}

	sorted := getSortedSemVerList(versions)
	for k, v := range sorted {
		if pass[k] != v.String() {
			t.Errorf("Sorting expected %s but got %s", pass[k], v)
		}
	}
}
