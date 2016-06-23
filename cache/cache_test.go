package cache

import "testing"

func TestKey(t *testing.T) {
	tests := map[string]string{
		"https://github.com/foo/bar":     "https-github.com-foo-bar",
		"git@github.com:foo/bar":         "git-github.com-foo-bar",
		"https://github.com:123/foo/bar": "https-github.com-123-foo-bar",
	}

	for k, v := range tests {
		key, err := Key(k)
		if err != nil {
			t.Errorf("Cache key generation err: %s", err)
			continue
		}
		if key != v {
			t.Errorf("Expected cache key %s for %s but got %s", v, k, key)
		}
	}
}
