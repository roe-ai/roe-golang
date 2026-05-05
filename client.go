package roe

import (
	"context"
	"fmt"
	"net/http"

	"github.com/roe-ai/roe-golang/generated"
)

// RoeClient is the main entrypoint.
type RoeClient struct {
	Config Config
	auth   Auth
	http   *httpClient

	Agents   *AgentsAPI
	Policies *PoliciesAPI
}

// NewClient constructs a RoeClient using parameters or environment fallbacks.
func NewClient(apiKey, organizationID, baseURL string, timeoutSeconds float64, maxRetries int) (*RoeClient, error) {
	cfg, err := LoadConfig(apiKey, organizationID, baseURL, timeoutSeconds, maxRetries)
	if err != nil {
		return nil, err
	}
	return NewClientWithConfig(cfg)
}

// NewClientWithParams constructs a RoeClient from structured configuration parameters.
func NewClientWithParams(params ConfigParams) (*RoeClient, error) {
	cfg, err := LoadConfigWithParams(params)
	if err != nil {
		return nil, err
	}
	return NewClientWithConfig(cfg)
}

// NewClientWithConfig builds a RoeClient from a fully parsed Config.
func NewClientWithConfig(cfg Config) (*RoeClient, error) {
	auth := newAuth(cfg)
	httpClient := newHTTPClient(cfg, auth)
	agentsAPI := newAgentsAPI(cfg, httpClient)
	policiesAPI := newPoliciesAPI(cfg, httpClient)

	return &RoeClient{
		Config:   cfg,
		auth:     auth,
		http:     httpClient,
		Agents:   agentsAPI,
		Policies: policiesAPI,
	}, nil
}

// Close releases HTTP resources.
func (c *RoeClient) Close() {
	if c == nil || c.http == nil {
		return
	}
	c.http.close()
}

// Raw returns the generated OpenAPI client configured with the same base URL,
// auth headers, and underlying http.Client as the ergonomic SDK surface.
func (c *RoeClient) Raw(opts ...generated.ClientOption) (*generated.ClientWithResponses, error) {
	if c == nil || c.http == nil || c.http.client == nil {
		return nil, fmt.Errorf("roe client is not initialized")
	}

	options := []generated.ClientOption{
		generated.WithHTTPClient(c.http.client),
		generated.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			for key, values := range c.auth.Headers() {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
			for key, values := range c.Config.ExtraHeaders {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
			for _, hook := range c.Config.BeforeRequest {
				hook(req)
			}
			return ctx.Err()
		}),
	}
	options = append(options, opts...)
	return generated.NewClientWithResponses(c.Config.BaseURL, options...)
}
