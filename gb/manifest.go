// Package gb provides compatibility with GB manifests.
package gb

// This is lifted wholesale from GB's `vendor/manifest.go` file.
//
// gb's license is MIT-style.

// Manifest represents the GB manifest file
type Manifest struct {
	Version      int          `json:"version"`
	Dependencies []Dependency `json:"dependencies"`
}

// Dependency represents an individual dependency in the GB manifest file
type Dependency struct {
	Importpath string `json:"importpath"`
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Branch     string `json:"branch"`
	Path       string `json:"path,omitempty"`
}
