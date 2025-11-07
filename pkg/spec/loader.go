package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing OpenAPI specifications
type Loader struct {
	loader *openapi3.Loader
}

// NewLoader creates a new spec loader
func NewLoader() *Loader {
	return &Loader{
		loader: openapi3.NewLoader(),
	}
}

// LoadFromFile loads an OpenAPI or Swagger spec from a file path
func (l *Loader) LoadFromFile(ctx context.Context, path string) (*openapi3.T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	return l.loadFromData(ctx, data)
}

// LoadFromURL loads an OpenAPI or Swagger spec from a URL
func (l *Loader) LoadFromURL(ctx context.Context, specURL string) (*openapi3.T, error) {
	resp, err := http.Get(specURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch spec: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec: %w", err)
	}

	return l.loadFromData(ctx, data)
}

// loadFromData loads spec from raw data, detecting and converting Swagger 2.0 if needed
func (l *Loader) loadFromData(ctx context.Context, data []byte) (*openapi3.T, error) {
	var rawMap map[string]interface{}

	if err := json.Unmarshal(data, &rawMap); err != nil {
		if err := yaml.Unmarshal(data, &rawMap); err != nil {
			return nil, fmt.Errorf("failed to parse spec as JSON or YAML: %w", err)
		}
	}

	if swagger, ok := rawMap["swagger"].(string); ok && strings.HasPrefix(swagger, "2.") {
		return l.loadSwagger2(data)
	}

	doc, err := l.loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("OpenAPI spec validation failed: %w", err)
	}

	return doc, nil
}

// loadSwagger2 loads and converts Swagger 2.0 spec to OpenAPI 3.0
func (l *Loader) loadSwagger2(data []byte) (*openapi3.T, error) {
	var rawData interface{}

	if err := json.Unmarshal(data, &rawData); err != nil {
		if err := yaml.Unmarshal(data, &rawData); err != nil {
			return nil, fmt.Errorf("failed to parse spec: %w", err)
		}

		rawData = convertYAMLMapToJSON(rawData)

		var err error
		data, err = json.Marshal(rawData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
		}
	}

	var v2 openapi2.T
	if err := json.Unmarshal(data, &v2); err != nil {
		return nil, fmt.Errorf("failed to parse Swagger 2.0 spec: %w", err)
	}

	v3, err := openapi2conv.ToV3(&v2)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Swagger 2.0 to OpenAPI 3.0: %w", err)
	}

	return v3, nil
}

// convertYAMLMapToJSON converts map[interface{}]interface{} to map[string]interface{}
// This is necessary because YAML unmarshaling creates maps with interface{} keys,
// but JSON requires string keys
func convertYAMLMapToJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			strKey := fmt.Sprintf("%v", key)
			result[strKey] = convertYAMLMapToJSON(value)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = convertYAMLMapToJSON(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = convertYAMLMapToJSON(value)
		}
		return result
	default:
		return v
	}
}
