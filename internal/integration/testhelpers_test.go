package integration_test

import "github.com/sushichan044/ajisai/internal/domain"

// makeTestURI creates a test URI for use in test cases.
func makeTestURI(path string, itemType domain.PresetType) domain.URI {
	return domain.URI{
		Scheme:  domain.Scheme,
		Package: "test-package", // テストでは常に固定値
		Preset:  "test-preset",  // テストでは常に固定値
		Type:    itemType,
		Path:    path,
	}
}
