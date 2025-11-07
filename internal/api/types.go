package api

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// Spec wraps an OpenAPI specification with metadata
type Spec struct {
	Doc      *openapi3.T
	Source   string
	Title    string
	Version  string
	BaseURL  string
}

// Parameter represents a request parameter
type Parameter struct {
	Name        string
	In          string // path, query, header, cookie
	Required    bool
	Description string
	Schema      string
	Example     string
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Path         string
	Method       string
	Summary      string
	Description  string
	Tags         []string
	Parameters   []Parameter
	RequestBody  string
	HasBody      bool
}

// GetEndpoints extracts all endpoints from the spec
func (s *Spec) GetEndpoints() []Endpoint {
	var endpoints []Endpoint

	if s.Doc == nil || s.Doc.Paths == nil {
		return endpoints
	}

	for path, pathItem := range s.Doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			endpoint := Endpoint{
				Path:        path,
				Method:      method,
				Summary:     operation.Summary,
				Description: operation.Description,
				Tags:        operation.Tags,
				Parameters:  extractParameters(operation),
				HasBody:     operation.RequestBody != nil,
			}

			if operation.RequestBody != nil && operation.RequestBody.Value != nil {
				content := operation.RequestBody.Value.Content
				if jsonContent := content.Get("application/json"); jsonContent != nil && jsonContent.Example != nil {
					endpoint.RequestBody = formatExample(jsonContent.Example)
				}
			}

			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

func extractParameters(op *openapi3.Operation) []Parameter {
	var params []Parameter

	for _, paramRef := range op.Parameters {
		if paramRef.Value == nil {
			continue
		}

		p := paramRef.Value
		param := Parameter{
			Name:        p.Name,
			In:          p.In,
			Required:    p.Required,
			Description: p.Description,
		}

		if p.Schema != nil && p.Schema.Value != nil {
			param.Schema = p.Schema.Value.Type.Slice()[0]
			if p.Schema.Value.Example != nil {
				param.Example = formatExample(p.Schema.Value.Example)
			}
		}

		params = append(params, param)
	}

	return params
}

func formatExample(example interface{}) string {
	if example == nil {
		return ""
	}
	if s, ok := example.(string); ok {
		return s
	}
	// For complex types, return a simple string representation
	return fmt.Sprintf("%v", example)
}

// GetInfo returns basic spec information
func (s *Spec) GetInfo() (title, version, description string) {
	if s.Doc != nil && s.Doc.Info != nil {
		title = s.Doc.Info.Title
		version = s.Doc.Info.Version
		description = s.Doc.Info.Description
	}
	return
}
