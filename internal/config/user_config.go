package config

type UserTomlConfig struct {
	Global  UserTomlGlobalConfig            `toml:"global,omitempty"`
	Inputs  map[string]UserTomlInputSource  `toml:"inputs,omitempty"`
	Outputs map[string]UserTomlOutputTarget `toml:"outputs,omitempty"`
}

type UserTomlGlobalConfig struct {
	CacheDir  string `toml:"cacheDir,omitempty"`
	Namespace string `toml:"namespace,omitempty"`
}

type UserTomlInputSource struct {
	Type       string `toml:"type"`                 // Required
	Path       string `toml:"path,omitempty"`       // Used if type=local
	Repository string `toml:"repository,omitempty"` // Used if type=git
	Revision   string `toml:"revision,omitempty"`   // Used if type=git (Optional ref/branch/tag/commit)
	SubDir     string `toml:"subDir,omitempty"`     // Used if type=git (Optional)
}

type UserTomlOutputTarget struct {
	Target  string `toml:"target"`
	Enabled bool   `toml:"enabled"`
}
