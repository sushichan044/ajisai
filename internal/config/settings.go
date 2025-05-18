package config

type Settings struct {
	// Specifies the directory where `ajisai` will store cached data of imported
	// presets.
	CacheDir string

	// Whether to enable experimental features.
	Experimental bool

	// A namespace string that can be used by output targets to organize or prefix the
	// imported presets.
	// For example, ajisai might place presets under `.cursor/prompts/<namespace>/` or
	// `.cursor/rules/<namespace>/`
	Namespace string
}

func applyDefaultsToSettings(settings *Settings) *Settings {
	if settings == nil {
		settings = &Settings{}
	}

	if settings.CacheDir == "" {
		settings.CacheDir = "./.cache/ajisai"
	}

	if settings.Namespace == "" {
		settings.Namespace = "ajisai"
	}

	return settings
}
