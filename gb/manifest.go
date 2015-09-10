// Package gb provides compatibility with GB manifests.
package gb

// This is lifted wholesale from GB's `vendor/manifest.go` file.
//
// gb's license is MIT-style.

type Manifest struct {
	Version      int          `json:"version"`
	Dependencies []Dependency `json:"dependencies"`
}

type Dependency struct {
	Importpath string `json:"importpath"`
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Branch     string `json:"branch"`
	Path       string `json:"path,omitempty"`
}
