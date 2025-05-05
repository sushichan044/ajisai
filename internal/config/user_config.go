package config

// UserConfig maps directly to the ai-rules.toml file structure for parsing.
// Fields are typically pointers or use types allowing omitempty behavior.
type UserConfig struct {
	Global  *UserGlobalConfig           `toml:"global,omitempty"`
	Inputs  map[string]UserInputSource  `toml:"inputs,omitempty"`
	Outputs map[string]UserOutputTarget `toml:"outputs,omitempty"`
}

// UserGlobalConfig represents the optional [global] section in TOML.
type UserGlobalConfig struct {
	CacheDir  *string `toml:"cacheDir,omitempty"`
	Namespace *string `toml:"namespace,omitempty"`
}

// UserInputSource represents an entry in the [inputs] section in TOML.
// This struct captures all possible fields for TOML parsing.
// The ConfigManager.Load method converts this into the domain.InputSource,
// performing validation based on the 'Type' field (e.g., ensuring 'path' is present
// for type "local" and absent/ignored for type "git").
// Note: Due to the mixed optional fields based on 'Type', directly generating
// a strict schema (like JSON Schema) from this struct might be misleading.
// A custom schema definition or runtime validation is necessary to enforce
// type-specific field requirements.
type UserInputSource struct {
	Type       string  `toml:"type"`                 // Required
	Path       *string `toml:"path,omitempty"`       // Used if type=local
	Repository *string `toml:"repository,omitempty"` // Used if type=git
	Revision   *string `toml:"revision,omitempty"`   // Used if type=git (Optional ref/branch/tag/commit)
	SubDir     *string `toml:"subDir,omitempty"`     // Used if type=git (Optional)
	// Format     *string `toml:"format,omitempty"` // Optional, parser type - Currently unused
}

// UserOutputTarget represents an entry in the [outputs] section in TOML.
type UserOutputTarget struct {
	Target  string `toml:"target"`            // Required
	Enabled *bool  `toml:"enabled,omitempty"` // Optional, defaults to true if omitted
}
