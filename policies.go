package roe

import (
	"context"
	"encoding/json"
	"fmt"
)

// PoliciesAPI manages policy operations for agentic workflows.
type PoliciesAPI struct {
	cfg        Config
	httpClient *httpClient
	Versions   *PolicyVersionsAPI
}

func newPoliciesAPI(cfg Config, httpClient *httpClient) *PoliciesAPI {
	api := &PoliciesAPI{cfg: cfg, httpClient: httpClient}
	api.Versions = &PolicyVersionsAPI{policiesAPI: api}
	return api
}

// List returns paginated policies.
func (p *PoliciesAPI) List(page, pageSize int) (PaginatedResponse[Policy], error) {
	return p.ListWithContext(context.Background(), page, pageSize)
}

// ListWithContext returns paginated policies with a caller-supplied context.
func (p *PoliciesAPI) ListWithContext(ctx context.Context, page, pageSize int) (PaginatedResponse[Policy], error) {
	params := map[string]string{
		"organization_id": p.cfg.OrganizationID,
	}
	if page > 0 {
		params["page"] = fmt.Sprintf("%d", page)
	}
	if pageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", pageSize)
	}
	var resp PaginatedResponse[Policy]
	if err := p.httpClient.getWithContext(ctx, "/v1/policies/", params, &resp); err != nil {
		return PaginatedResponse[Policy]{}, err
	}
	return resp, nil
}

// Retrieve fetches a policy by ID.
func (p *PoliciesAPI) Retrieve(policyID string) (Policy, error) {
	return p.RetrieveWithContext(context.Background(), policyID)
}

// RetrieveWithContext fetches a policy by ID with a caller-supplied context.
func (p *PoliciesAPI) RetrieveWithContext(ctx context.Context, policyID string) (Policy, error) {
	if policyID == "" {
		return Policy{}, fmt.Errorf("policyID cannot be empty")
	}
	var resp Policy
	if err := p.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/policies/%s/", policyID), nil, &resp); err != nil {
		return Policy{}, err
	}
	return resp, nil
}

// Create creates a new policy with an initial version.
func (p *PoliciesAPI) Create(name string, content map[string]any, description string, versionName string) (Policy, error) {
	return p.CreateWithContext(context.Background(), name, content, description, versionName)
}

// CreateWithContext creates a new policy with a caller-supplied context.
func (p *PoliciesAPI) CreateWithContext(ctx context.Context, name string, content map[string]any, description string, versionName string) (Policy, error) {
	payload := map[string]any{
		"name":            name,
		"content":         content,
		"description":     description,
		"organization_id": p.cfg.OrganizationID,
	}
	if versionName != "" {
		payload["version_name"] = versionName
	}
	var resp Policy
	if err := p.httpClient.postJSONWithContext(ctx, "/v1/policies/", payload, nil, &resp); err != nil {
		return Policy{}, err
	}
	return resp, nil
}

// Update updates a policy's metadata. Pass nil for fields you don't want to change.
func (p *PoliciesAPI) Update(policyID string, name *string, description *string) (Policy, error) {
	return p.UpdateWithContext(context.Background(), policyID, name, description)
}

// UpdateWithContext updates a policy with a caller-supplied context.
func (p *PoliciesAPI) UpdateWithContext(ctx context.Context, policyID string, name *string, description *string) (Policy, error) {
	if policyID == "" {
		return Policy{}, fmt.Errorf("policyID cannot be empty")
	}
	payload := map[string]any{}
	if name != nil {
		payload["name"] = *name
	}
	if description != nil {
		payload["description"] = *description
	}
	var resp Policy
	if err := p.httpClient.putJSONWithContext(ctx, fmt.Sprintf("/v1/policies/%s/", policyID), payload, nil, &resp); err != nil {
		return Policy{}, err
	}
	return resp, nil
}

// Delete removes a policy and all its versions.
func (p *PoliciesAPI) Delete(policyID string) error {
	return p.DeleteWithContext(context.Background(), policyID)
}

// DeleteWithContext removes a policy with a caller-supplied context.
func (p *PoliciesAPI) DeleteWithContext(ctx context.Context, policyID string) error {
	if policyID == "" {
		return fmt.Errorf("policyID cannot be empty")
	}
	return p.httpClient.deleteWithContext(ctx, fmt.Sprintf("/v1/policies/%s/", policyID), nil)
}

// PolicyVersionsAPI handles policy version operations.
type PolicyVersionsAPI struct {
	policiesAPI *PoliciesAPI
}

// List returns all versions of a policy.
func (v *PolicyVersionsAPI) List(policyID string) ([]PolicyVersion, error) {
	return v.ListWithContext(context.Background(), policyID)
}

// ListWithContext returns all versions of a policy with a caller-supplied context.
func (v *PolicyVersionsAPI) ListWithContext(ctx context.Context, policyID string) ([]PolicyVersion, error) {
	if policyID == "" {
		return nil, fmt.Errorf("policyID cannot be empty")
	}
	// The endpoint may return either a paginated response or a raw array.
	raw, err := v.policiesAPI.httpClient.getBytesWithContext(ctx, fmt.Sprintf("/v1/policies/%s/versions/", policyID), nil)
	if err != nil {
		return nil, err
	}
	// Try paginated format first.
	var paginated struct {
		Results []PolicyVersion `json:"results"`
	}
	if len(raw) > 0 && raw[0] == '{' {
		if err := json.Unmarshal(raw, &paginated); err != nil {
			return nil, fmt.Errorf("parse policy versions response: %w", err)
		}
		return paginated.Results, nil
	}
	// Fall back to raw array.
	var versions []PolicyVersion
	if err := json.Unmarshal(raw, &versions); err != nil {
		return nil, fmt.Errorf("parse policy versions response: %w", err)
	}
	return versions, nil
}

// Retrieve fetches a specific policy version.
func (v *PolicyVersionsAPI) Retrieve(policyID, versionID string) (PolicyVersion, error) {
	return v.RetrieveWithContext(context.Background(), policyID, versionID)
}

// RetrieveWithContext fetches a specific policy version with a caller-supplied context.
func (v *PolicyVersionsAPI) RetrieveWithContext(ctx context.Context, policyID, versionID string) (PolicyVersion, error) {
	if policyID == "" {
		return PolicyVersion{}, fmt.Errorf("policyID cannot be empty")
	}
	if versionID == "" {
		return PolicyVersion{}, fmt.Errorf("versionID cannot be empty")
	}
	var resp PolicyVersion
	if err := v.policiesAPI.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/policies/%s/versions/%s/", policyID, versionID), nil, &resp); err != nil {
		return PolicyVersion{}, err
	}
	return resp, nil
}

// Create creates a new policy version. The new version automatically becomes current.
func (v *PolicyVersionsAPI) Create(policyID string, content map[string]any, versionName string, baseVersionID string) (PolicyVersion, error) {
	return v.CreateWithContext(context.Background(), policyID, content, versionName, baseVersionID)
}

// CreateWithContext creates a new policy version with a caller-supplied context.
func (v *PolicyVersionsAPI) CreateWithContext(ctx context.Context, policyID string, content map[string]any, versionName string, baseVersionID string) (PolicyVersion, error) {
	if policyID == "" {
		return PolicyVersion{}, fmt.Errorf("policyID cannot be empty")
	}
	payload := map[string]any{
		"content": content,
	}
	if versionName != "" {
		payload["version_name"] = versionName
	}
	if baseVersionID != "" {
		payload["base_version_id"] = baseVersionID
	}
	// POST returns partial data; extract ID and re-fetch.
	var respID struct {
		ID string `json:"id"`
	}
	if err := v.policiesAPI.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/policies/%s/versions/", policyID), payload, nil, &respID); err != nil {
		return PolicyVersion{}, err
	}
	if respID.ID == "" {
		return PolicyVersion{}, fmt.Errorf("unexpected response: missing version ID")
	}
	return v.RetrieveWithContext(ctx, policyID, respID.ID)
}
