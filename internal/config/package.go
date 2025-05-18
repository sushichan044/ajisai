package config

type (
	Package struct {
		// Exported preset definitions.
		//
		// Key is the exported preset name.
		Exports map[string]ExportedPresetDefinition `json:"exports,omitempty"`

		// Package name.
		Name string `json:"name"`
	}

	ExportedPresetDefinition struct {
		// Prompts to export.
		//
		// You can use glob patterns supported by
		// [bmatcuk/doublestart](https://github.com/bmatcuk/doublestar)
		Prompts []string `json:"prompts,omitempty"`

		// Rules to export.
		//
		// You can use glob patterns supported by
		// [bmatcuk/doublestart](https://github.com/bmatcuk/doublestar)
		Rules []string `json:"rules,omitempty"`
	}
)

func applyDefaultsToPackage(pkg *Package) *Package {
	if pkg == nil {
		pkg = &Package{}
	}

	if pkg.Exports == nil {
		pkg.Exports = map[string]ExportedPresetDefinition{}
	}

	return pkg
}
