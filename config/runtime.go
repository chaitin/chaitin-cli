package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Raw map[string]yaml.Node

func LoadEnvFile(path string) error {
	if path == "" {
		return nil
	}

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat env file %s: %w", path, err)
	}

	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("load env file %s: %w", path, err)
	}

	return nil
}

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
		applyEnvOverrides(name, &cfg)
		return cfg, nil
	}

	node, ok := r[name]
	if !ok {
		applyEnvOverrides(name, &cfg)
		return cfg, nil
	}

	if err := node.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode config for %s: %w", name, err)
	}

	applyEnvOverrides(name, &cfg)

	return cfg, nil
}

func applyEnvOverrides(product string, target any) {
	value := reflect.ValueOf(target)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return
	}

	envPrefix := normalizeEnvSegment(product)
	applyStructEnvOverrides(elem, []string{envPrefix})
}

func applyStructEnvOverrides(v reflect.Value, path []string) {
	t := v.Type()
	for i := range t.NumField() {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		tagName, ok := yamlFieldName(fieldType)
		if !ok {
			continue
		}

		nextPath := append(path, normalizeEnvSegment(tagName))
		switch fieldValue.Kind() {
		case reflect.Struct:
			applyStructEnvOverrides(fieldValue, nextPath)
		case reflect.Pointer:
			if fieldValue.Type().Elem().Kind() != reflect.Struct {
				continue
			}
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			applyStructEnvOverrides(fieldValue.Elem(), nextPath)
		default:
			raw, ok := os.LookupEnv(strings.Join(nextPath, "_"))
			if !ok || raw == "" {
				continue
			}

			if err := decodeScalar(raw, fieldValue); err != nil {
				continue
			}
		}
	}
}

func yamlFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("yaml")
	if tag == "-" {
		return "", false
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		name = field.Name
	}
	if name == "-" {
		return "", false
	}
	return name, true
}

// normalizeEnvSegment converts a config key into its environment-variable
// segment. Dashes and dots become underscores, and camelCase boundaries are
// split so that e.g. "cloudAtlas" becomes "CLOUD_ATLAS".
func normalizeEnvSegment(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")

	var b strings.Builder
	var prev rune
	for _, r := range s {
		if unicode.IsUpper(r) && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
			b.WriteRune('_')
		}
		b.WriteRune(r)
		prev = r
	}
	return strings.ToUpper(b.String())
}

func decodeScalar(raw string, target reflect.Value) error {
	node := yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: raw,
	}
	return node.Decode(target.Addr().Interface())
}
