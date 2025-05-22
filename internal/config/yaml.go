package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"

	"github.com/sushichan044/ajisai/utils"
)

type yamlLoader struct {
	serializer configSerializer
}

func newYAMLLoader() formatLoader[SerializableConfig] {
	return &yamlLoader{
		serializer: NewSerializer(),
	}
}

func (l *yamlLoader) Load(configPath string) (*Config, error) {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	if _, statErr := os.Stat(resolvedPath); statErr != nil {
		return nil, fmt.Errorf("failed to get config file %s: %w", resolvedPath, statErr)
	}

	body, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", resolvedPath, err)
	}

	var cfgData SerializableConfig
	if yamlErr := yaml.Unmarshal(body, &cfgData); yamlErr != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %s: %w", resolvedPath, yamlErr)
	}

	return l.serializer.Deserialize(cfgData)
}

func (l *yamlLoader) Save(configPath string, cfg *Config) error {
	resolvedPath, pathErr := utils.ResolveAbsPath(configPath)
	if pathErr != nil {
		return fmt.Errorf("failed to resolve config path: %w", pathErr)
	}
	cfgData, convErr := l.serializer.Serialize(cfg)
	if convErr != nil {
		return fmt.Errorf("failed to convert config to serializable format: %w", convErr)
	}

	yamlData, marshalErr := yaml.Marshal(cfgData)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", marshalErr)
	}

	if err := utils.AtomicWriteFile(resolvedPath, bytes.NewReader(yamlData)); err != nil {
		return fmt.Errorf("failed to save config file atomically: %w", err)
	}

	return nil
}
