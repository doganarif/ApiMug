package api

import (
	"fmt"
	"net/http"
)

// AuthType represents the type of authentication
type AuthType string

const (
	AuthTypeNone       AuthType = "none"
	AuthTypeBearer     AuthType = "bearer"
	AuthTypeAPIKey     AuthType = "apikey"
	AuthTypeBasic      AuthType = "basic"
	AuthTypeOAuth2     AuthType = "oauth2"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type     AuthType
	Token    string // For bearer tokens
	APIKey   string // For API keys
	APIKeyIn string // header, query, cookie
	KeyName  string // Name of the header/query param
	Username string // For basic auth
	Password string // For basic auth
}

// AuthManager manages authentication for API requests
type AuthManager struct {
	config *AuthConfig
	spec   *Spec
}

// NewAuthManager creates a new auth manager
func NewAuthManager(spec *Spec) *AuthManager {
	return &AuthManager{
		spec:   spec,
		config: &AuthConfig{Type: AuthTypeNone},
	}
}

// GetAvailableAuthSchemes returns available auth schemes from the spec
func (am *AuthManager) GetAvailableAuthSchemes() []string {
	if am.spec.Doc == nil || am.spec.Doc.Components == nil || am.spec.Doc.Components.SecuritySchemes == nil {
		return []string{"none"}
	}

	schemes := []string{"none"}
	for name := range am.spec.Doc.Components.SecuritySchemes {
		schemes = append(schemes, name)
	}
	return schemes
}

// SetAuth configures authentication
func (am *AuthManager) SetAuth(config *AuthConfig) {
	am.config = config
}

// GetAuth returns current auth config
func (am *AuthManager) GetAuth() *AuthConfig {
	return am.config
}

// ApplyAuth applies authentication to an HTTP request
func (am *AuthManager) ApplyAuth(req *http.Request) error {
	if am.config == nil || am.config.Type == AuthTypeNone {
		return nil
	}

	switch am.config.Type {
	case AuthTypeBearer:
		if am.config.Token == "" {
			return fmt.Errorf("bearer token is required")
		}
		req.Header.Set("Authorization", "Bearer "+am.config.Token)

	case AuthTypeAPIKey:
		if am.config.APIKey == "" {
			return fmt.Errorf("API key is required")
		}
		switch am.config.APIKeyIn {
		case "header":
			req.Header.Set(am.config.KeyName, am.config.APIKey)
		case "query":
			q := req.URL.Query()
			q.Set(am.config.KeyName, am.config.APIKey)
			req.URL.RawQuery = q.Encode()
		case "cookie":
			req.AddCookie(&http.Cookie{
				Name:  am.config.KeyName,
				Value: am.config.APIKey,
			})
		}

	case AuthTypeBasic:
		if am.config.Username == "" || am.config.Password == "" {
			return fmt.Errorf("username and password are required")
		}
		req.SetBasicAuth(am.config.Username, am.config.Password)

	case AuthTypeOAuth2:
		if am.config.Token == "" {
			return fmt.Errorf("OAuth2 token is required")
		}
		req.Header.Set("Authorization", "Bearer "+am.config.Token)
	}

	return nil
}

// ParseAuthScheme parses an auth scheme from the OpenAPI spec
func (am *AuthManager) ParseAuthScheme(schemeName string) (*AuthConfig, error) {
	if schemeName == "none" {
		return &AuthConfig{Type: AuthTypeNone}, nil
	}

	if am.spec.Doc == nil || am.spec.Doc.Components == nil || am.spec.Doc.Components.SecuritySchemes == nil {
		return nil, fmt.Errorf("no security schemes defined")
	}

	schemeRef := am.spec.Doc.Components.SecuritySchemes[schemeName]
	if schemeRef == nil {
		return nil, fmt.Errorf("security scheme %s not found", schemeName)
	}

	scheme := schemeRef.Value
	config := &AuthConfig{}

	switch scheme.Type {
	case "http":
		if scheme.Scheme == "bearer" {
			config.Type = AuthTypeBearer
		} else if scheme.Scheme == "basic" {
			config.Type = AuthTypeBasic
		}

	case "apiKey":
		config.Type = AuthTypeAPIKey
		config.KeyName = scheme.Name
		config.APIKeyIn = scheme.In

	case "oauth2":
		config.Type = AuthTypeOAuth2

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", scheme.Type)
	}

	return config, nil
}
