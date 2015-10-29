package cmd

import "testing"

func TestCacheCreateKey(t *testing.T) {
	tests := map[string]string{
		"https://github.com/foo/bar": "https-github.com-foo-bar",
		"git@github.com:foo/bar":     "git-github.com-foo-bar",
	}

	for k, v := range tests {
		key, err := cacheCreateKey(k)
		if err != nil {
			t.Errorf("Cache key generation err: %s", err)
			continue
		}
		if key != v {
			t.Errorf("Expected cache key %s for %s but got %s", v, k, key)
		}
	}
}
