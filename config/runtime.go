package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Raw map[string]yaml.Node

func Load(path string) (Raw, error) {
	cfg := Raw{}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}

	return cfg, nil
}

func DecodeProduct[T any](r Raw, name string) (T, error) {
	var cfg T
	if r == nil {
		return cfg, nil
	}

	node, ok := r[name]
	if !ok {
		return cfg, nil
	}

	if err := node.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode config for %s: %w", name, err)
	}

	return cfg, nil
}
