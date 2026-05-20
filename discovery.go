package roe

import (
	"context"
)

// DiscoveryAPI exposes metadata needed to choose engine_class_id and model IDs.
type DiscoveryAPI struct {
	cfg        Config
	httpClient *httpClient
}

func newDiscoveryAPI(cfg Config, httpClient *httpClient) *DiscoveryAPI {
	return &DiscoveryAPI{cfg: cfg, httpClient: httpClient}
}

// ListAgentEngineTypes returns production engine_class_id values accepted by agent creation.
func (d *DiscoveryAPI) ListAgentEngineTypes() (AgentEngineTypeList, error) {
	return d.ListAgentEngineTypesWithContext(context.Background())
}

// ListAgentEngineTypesWithContext returns agent engine types with a caller-supplied context.
func (d *DiscoveryAPI) ListAgentEngineTypesWithContext(ctx context.Context) (AgentEngineTypeList, error) {
	var resp AgentEngineTypeList
	if err := d.httpClient.getWithContext(ctx, "/v1/agents/types/", nil, &resp); err != nil {
		return AgentEngineTypeList{}, err
	}
	return resp, nil
}

// ListSupportedModels returns non-deprecated model IDs accepted in engine_config.model.
func (d *DiscoveryAPI) ListSupportedModels(capability string) (SupportedLLMModelList, error) {
	return d.ListSupportedModelsWithContext(context.Background(), capability)
}

// ListSupportedModelsWithContext returns supported models with a caller-supplied context.
func (d *DiscoveryAPI) ListSupportedModelsWithContext(ctx context.Context, capability string) (SupportedLLMModelList, error) {
	query := map[string]string{}
	if capability != "" {
		query["capability"] = capability
	}
	var resp SupportedLLMModelList
	if err := d.httpClient.getWithContext(ctx, "/v1/agents/models/", query, &resp); err != nil {
		return SupportedLLMModelList{}, err
	}
	return resp, nil
}
