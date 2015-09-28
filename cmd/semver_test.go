package cmd

import (
	"testing"
)

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
