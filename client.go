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
	raw    *generated.ClientWithResponses

	Agents   *AgentsAPI
	Policies *PoliciesAPI
	Users    *UsersAPI
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

	client := &RoeClient{
		Config: cfg,
		auth:   auth,
		http:   httpClient,
	}

	raw, err := buildRawClient(client)
	if err != nil {
		return nil, err
	}
	client.raw = raw

	client.Agents = newAgentsAPI(client)
	client.Policies = newPoliciesAPI(client)
	client.Users = newUsersAPI(client)

	return client, nil
}

// Close releases HTTP resources.
func (c *RoeClient) Close() {
	if c == nil || c.http == nil {
		return
	}
	c.http.close()
}

// Raw returns the generated OpenAPI client configured with the same base URL,
// auth headers, retry policy, and request hooks as the ergonomic SDK surface.
//
// Without options, the cached client is returned. When custom ClientOptions
// are supplied, a fresh ClientWithResponses is built so caller overrides do
// not leak into the cached instance shared by the wrappers.
func (c *RoeClient) Raw(opts ...generated.ClientOption) (*generated.ClientWithResponses, error) {
	if c == nil || c.http == nil || c.http.client == nil {
		return nil, fmt.Errorf("roe client is not initialized")
	}
	if len(opts) == 0 && c.raw != nil {
		return c.raw, nil
	}
	return buildRawClient(c, opts...)
}

func buildRawClient(c *RoeClient, extra ...generated.ClientOption) (*generated.ClientWithResponses, error) {
	options := []generated.ClientOption{
		generated.WithHTTPClient(&retryDoer{c: c.http}),
		generated.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			// Apply auth + static extra headers here so the generated
			// client's request carries them before the retryDoer hands
			// off to httpClient.doRetried (which mirrors the legacy
			// doRequest -> applyHeaders behaviour).
			//
			// BeforeRequest hooks are intentionally NOT invoked here:
			// doRetried calls runRequestHooks per attempt, so running
			// them in the editor would double-fire on the first attempt
			// (and miss the per-retry replay semantics callers expect
			// from the legacy path). Hooks fire exactly once per HTTP
			// attempt regardless of whether the request originates from
			// the generated client or the legacy hand-written helpers.
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
			return ctx.Err()
		}),
	}
	options = append(options, extra...)
	return generated.NewClientWithResponses(c.Config.BaseURL, options...)
}
