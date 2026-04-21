package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func SetProduct(path, name string, value any) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}

	var node yaml.Node
	if err := node.Encode(value); err != nil {
		return fmt.Errorf("encode config for %s: %w", name, err)
	}
	cfg[name] = node

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config file %s: %w", path, err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config file %s: %w", path, err)
	}

	return nil
}
