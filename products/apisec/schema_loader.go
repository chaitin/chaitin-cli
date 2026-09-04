package apisec

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed latest v26.05/openapi.json v26.05/cli-mapping.yaml
var schemaFS embed.FS

func parseOpenAPI(data []byte) (*OpenAPI, error) {
	var api OpenAPI
	if err := json.Unmarshal(data, &api); err != nil {
		return nil, fmt.Errorf("parse OpenAPI: %w", err)
	}
	return &api, nil
}

func parseCLIMapping(data []byte) (*CLIMapping, error) {
	var mapping CLIMapping
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("parse CLI mapping: %w", err)
	}
	return &mapping, nil
}

func loadEmbeddedSchema() (*OpenAPI, *CLIMapping, error) {
	versionData, err := schemaFS.ReadFile("latest")
	if err != nil {
		return nil, nil, fmt.Errorf("read latest APISec schema version: %w", err)
	}
	version := strings.TrimSpace(string(versionData))
	if version == "" {
		return nil, nil, fmt.Errorf("latest APISec schema version is empty")
	}

	openAPIData, err := schemaFS.ReadFile(path.Join(version, "openapi.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("read APISec OpenAPI for %s: %w", version, err)
	}
	api, err := parseOpenAPI(openAPIData)
	if err != nil {
		return nil, nil, err
	}

	mappingData, err := schemaFS.ReadFile(path.Join(version, "cli-mapping.yaml"))
	if err != nil {
		return nil, nil, fmt.Errorf("read APISec CLI mapping for %s: %w", version, err)
	}
	mapping, err := parseCLIMapping(mappingData)
	if err != nil {
		return nil, nil, err
	}

	return api, mapping, nil
}
