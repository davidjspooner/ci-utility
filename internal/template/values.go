package template

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Values holds a map of string key-value pairs loaded from a YAML file.
type Values struct {
	Values map[string]string `yaml:"values"`
}

// loadValues loads key-value pairs from a YAML file at the given path into a Values struct.
func loadValues(path string) (*Values, error) {
	data, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer data.Close()
	var values Values
	// Decode YAML into the Values struct.
	err = yaml.NewDecoder(data).Decode(&values.Values)
	if err != nil {
		return nil, err
	}
	return &values, nil
}

// Get retrieves the value associated with the given key from the Values struct.
func (v *Values) Get(key string) (string, error) {
	if value, ok := v.Values[key]; ok {
		return value, nil
	}
	return "", fmt.Errorf("key %s not found in values", key)
}

// Functions returns a map of template functions available for use in templates.
func (v *Values) Functions() map[string]any {
	templateFunctions := map[string]any{
		"env": func(key string) string {
			// Lookup environment variable by key.
			return os.Getenv(key)
		},
		"file": func(path string) (string, error) {
			// Read file content as string.
			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("failed to read file %s: %w", path, err)
			}
			return string(data), nil
		},
		"value": func(path string) (string, error) {
			// Read file content as string (same as file).
			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("failed to read file %s: %w", path, err)
			}
			return string(data), nil
		},
	}
	return templateFunctions
}
