package config

type (
	Package struct {
		// Exported preset definitions.
		Exports []ExportedPresetDefinition `json:"exports,omitempty"`

		// Package name.
		Name string `json:"name"`
	}

	ExportedPresetDefinition struct {
		// Preset name.
		Name string `json:"name"`

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
		pkg.Exports = []ExportedPresetDefinition{}
	}

	return pkg
}
