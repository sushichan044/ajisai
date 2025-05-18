package config

type (
	Package struct {
		// Exported preset definitions.
		//
		// Key is the exported preset name.
		Exports map[string]ExportedPresetDefinition

		// Package name.
		Name string
	}

	ExportedPresetDefinition struct {
		// Prompts to export.
		//
		// You can use glob patterns supported by
		// [bmatcuk/doublestart](https://github.com/bmatcuk/doublestar)
		Prompts []string

		// Rules to export.
		//
		// You can use glob patterns supported by
		// [bmatcuk/doublestart](https://github.com/bmatcuk/doublestar)
		Rules []string
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
