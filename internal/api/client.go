package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Request represents an API request
type Request struct {
	Method      string
	Path        string
	Headers     map[string]string
	QueryParams map[string]string
	Body        string
	ContentType string
}

// Response represents an API response
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       string
	Duration   time.Duration
	Error      error
}

// Client handles HTTP requests to the API
type Client struct {
	baseURL    string
	httpClient *http.Client
	authMgr    *AuthManager
}

// NewClient creates a new API client
func NewClient(baseURL string, authMgr *AuthManager) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authMgr: authMgr,
	}
}

// Send sends an HTTP request
func (c *Client) Send(ctx context.Context, req *Request) *Response {
	start := time.Now()
	resp := &Response{}

	// Build URL
	url := c.baseURL + req.Path
	if len(req.QueryParams) > 0 {
		url += "?"
		params := []string{}
		for k, v := range req.QueryParams {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		url += strings.Join(params, "&")
	}

	// Create request body
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		resp.Error = fmt.Errorf("failed to create request: %w", err)
		return resp
	}

	// Apply headers
	if req.ContentType != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Apply authentication
	if c.authMgr != nil {
		if err := c.authMgr.ApplyAuth(httpReq); err != nil {
			resp.Error = fmt.Errorf("auth error: %w", err)
			return resp
		}
	}

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		resp.Error = fmt.Errorf("request failed: %w", err)
		resp.Duration = time.Since(start)
		return resp
	}
	defer httpResp.Body.Close()

	// Read response
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Error = fmt.Errorf("failed to read response: %w", err)
	} else {
		resp.Body = string(bodyBytes)
	}

	resp.StatusCode = httpResp.StatusCode
	resp.Status = httpResp.Status
	resp.Headers = httpResp.Header
	resp.Duration = time.Since(start)

	return resp
}

// FormatResponseBody formats the response body for display
func (r *Response) FormatResponseBody() string {
	if r.Body == "" {
		return "(empty)"
	}

	// Try to pretty-print JSON
	contentType := r.Headers.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var parsed interface{}
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			if formatted, err := json.MarshalIndent(parsed, "", "  "); err == nil {
				return string(formatted)
			}
		}
	}

	return r.Body
}
