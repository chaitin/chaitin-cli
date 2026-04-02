package main

import (
	"path/filepath"

	"github.com/chaitin/workspace-cli/config"
)

const defaultConfigPath = "config.yaml"

func loadConfigFile(path string) (config.Raw, error) {
	return config.Load(path)
}

func configPathFromCWD() string {
	return filepath.Join(".", defaultConfigPath)
}
