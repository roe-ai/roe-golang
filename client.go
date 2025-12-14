package roe

// RoeClient is the main entrypoint.
type RoeClient struct {
	Config Config
	auth   Auth
	http   *httpClient

	Agents *AgentsAPI
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

	return &RoeClient{
		Config: cfg,
		auth:   auth,
		http:   httpClient,
		Agents: agentsAPI,
	}, nil
}

// Close releases HTTP resources.
func (c *RoeClient) Close() {
	if c == nil || c.http == nil {
		return
	}
	c.http.close()
}
