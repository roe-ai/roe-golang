package roe

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/roe-ai/roe-golang/generated"
)

// PoliciesAPI manages policy operations for agentic workflows. All methods
// delegate to the generated raw client; the wrapper layer provides ergonomic
// argument coercion (string IDs → UUID, optional pointers) and translates
// non-2xx responses to the typed RoeAPIException family via errorFromResponse.
type PoliciesAPI struct {
	cfg      Config
	client   *RoeClient
	Versions *PolicyVersionsAPI
}

func newPoliciesAPI(client *RoeClient) *PoliciesAPI {
	api := &PoliciesAPI{cfg: client.Config, client: client}
	api.Versions = &PolicyVersionsAPI{policiesAPI: api}
	return api
}

func (p *PoliciesAPI) raw() *generated.ClientWithResponses { return p.client.raw }

// orgUUID returns the configured organization ID parsed as a UUID, or nil if
// the SDK was constructed without one. Callers pass it through to generated
// *Params structs whose OrganizationId field is *openapi_types.UUID.
func (p *PoliciesAPI) orgUUID() (*openapi_types.UUID, error) {
	if p.cfg.OrganizationID == "" {
		return nil, nil
	}
	id, err := uuid.Parse(p.cfg.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	return &id, nil
}

func parseUUIDParam(name, value string) (openapi_types.UUID, error) {
	if value == "" {
		return openapi_types.UUID{}, fmt.Errorf("%s cannot be empty", name)
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return openapi_types.UUID{}, fmt.Errorf("invalid %s: %w", name, err)
	}
	return id, nil
}

// List returns paginated policies.
func (p *PoliciesAPI) List(page, pageSize int) (*generated.PaginatedPolicyList, error) {
	return p.ListWithContext(context.Background(), page, pageSize)
}

// ListWithContext returns paginated policies with a caller-supplied context.
func (p *PoliciesAPI) ListWithContext(ctx context.Context, page, pageSize int) (*generated.PaginatedPolicyList, error) {
	orgID, err := p.orgUUID()
	if err != nil {
		return nil, err
	}
	params := &generated.V1PoliciesListParams{OrganizationId: orgID}
	if page > 0 {
		params.Page = &page
	}
	if pageSize > 0 {
		params.PageSize = &pageSize
	}
	resp, err := p.raw().V1PoliciesListWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Retrieve fetches a policy by ID.
func (p *PoliciesAPI) Retrieve(policyID string) (*generated.Policy, error) {
	return p.RetrieveWithContext(context.Background(), policyID)
}

// RetrieveWithContext fetches a policy by ID with a caller-supplied context.
func (p *PoliciesAPI) RetrieveWithContext(ctx context.Context, policyID string) (*generated.Policy, error) {
	id, err := parseUUIDParam("policyID", policyID)
	if err != nil {
		return nil, err
	}
	orgID, err := p.orgUUID()
	if err != nil {
		return nil, err
	}
	resp, err := p.raw().V1PoliciesRetrieveWithResponse(ctx, id, &generated.V1PoliciesRetrieveParams{OrganizationId: orgID})
	if err != nil {
		return nil, err
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new policy with an initial version.
func (p *PoliciesAPI) Create(name string, content map[string]any, description string, versionName string) (*generated.CreatePolicy, error) {
	return p.CreateWithContext(context.Background(), name, content, description, versionName)
}

// CreateWithContext creates a new policy with a caller-supplied context.
func (p *PoliciesAPI) CreateWithContext(ctx context.Context, name string, content map[string]any, description string, versionName string) (*generated.CreatePolicy, error) {
	orgID, err := p.orgUUID()
	if err != nil {
		return nil, err
	}
	body := generated.CreatePolicyRequest{
		Name:    name,
		Content: content,
	}
	if description != "" {
		body.Description = &description
	}
	if versionName != "" {
		body.VersionName = &versionName
	}
	resp, err := p.raw().V1PoliciesCreateWithResponse(ctx, &generated.V1PoliciesCreateParams{OrganizationId: orgID}, body)
	if err != nil {
		return nil, err
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates a policy's metadata. Pass nil for fields you don't want to change.
// Uses PATCH semantics (only supplied fields are sent).
func (p *PoliciesAPI) Update(policyID string, name *string, description *string) (*generated.UpdatePolicy, error) {
	return p.UpdateWithContext(context.Background(), policyID, name, description)
}

// UpdateWithContext updates a policy with a caller-supplied context.
func (p *PoliciesAPI) UpdateWithContext(ctx context.Context, policyID string, name *string, description *string) (*generated.UpdatePolicy, error) {
	id, err := parseUUIDParam("policyID", policyID)
	if err != nil {
		return nil, err
	}
	orgID, err := p.orgUUID()
	if err != nil {
		return nil, err
	}
	body := generated.PatchedUpdatePolicyRequest{Name: name, Description: description}
	resp, err := p.raw().V1PoliciesPartialUpdateWithResponse(ctx, id, &generated.V1PoliciesPartialUpdateParams{OrganizationId: orgID}, body)
	if err != nil {
		return nil, err
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes a policy and all its versions.
func (p *PoliciesAPI) Delete(policyID string) error {
	return p.DeleteWithContext(context.Background(), policyID)
}

// DeleteWithContext removes a policy with a caller-supplied context.
func (p *PoliciesAPI) DeleteWithContext(ctx context.Context, policyID string) error {
	id, err := parseUUIDParam("policyID", policyID)
	if err != nil {
		return err
	}
	orgID, err := p.orgUUID()
	if err != nil {
		return err
	}
	resp, err := p.raw().V1PoliciesDestroyWithResponse(ctx, id, &generated.V1PoliciesDestroyParams{OrganizationId: orgID})
	if err != nil {
		return err
	}
	return errorFromResponse(resp.HTTPResponse, resp.Body)
}

// PolicyVersionsAPI handles policy version operations.
type PolicyVersionsAPI struct {
	policiesAPI *PoliciesAPI
}

// List returns all versions of a policy. The generated endpoint always
// returns a paginated response; the legacy "raw array" tolerance has been
// removed.
func (v *PolicyVersionsAPI) List(policyID string) (*generated.PaginatedPolicyVersionList, error) {
	return v.ListWithContext(context.Background(), policyID)
}

// ListWithContext returns all versions of a policy with a caller-supplied context.
func (v *PolicyVersionsAPI) ListWithContext(ctx context.Context, policyID string) (*generated.PaginatedPolicyVersionList, error) {
	id, err := parseUUIDParam("policyID", policyID)
	if err != nil {
		return nil, err
	}
	orgID, err := v.policiesAPI.orgUUID()
	if err != nil {
		return nil, err
	}
	resp, err := v.policiesAPI.raw().V1PoliciesVersionsListWithResponse(ctx, id, &generated.V1PoliciesVersionsListParams{OrganizationId: orgID})
	if err != nil {
		return nil, err
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Retrieve fetches a specific policy version.
func (v *PolicyVersionsAPI) Retrieve(policyID, versionID string) (*generated.PolicyVersion, error) {
	return v.RetrieveWithContext(context.Background(), policyID, versionID)
}

// RetrieveWithContext fetches a specific policy version with a caller-supplied context.
func (v *PolicyVersionsAPI) RetrieveWithContext(ctx context.Context, policyID, versionID string) (*generated.PolicyVersion, error) {
	pid, err := parseUUIDParam("policyID", policyID)
	if err != nil {
		return nil, err
	}
	vid, err := parseUUIDParam("versionID", versionID)
	if err != nil {
		return nil, err
	}
	orgID, err := v.policiesAPI.orgUUID()
	if err != nil {
		return nil, err
	}
	resp, err := v.policiesAPI.raw().V1PoliciesVersionsRetrieveWithResponse(ctx, pid, vid, &generated.V1PoliciesVersionsRetrieveParams{OrganizationId: orgID})
	if err != nil {
		return nil, err
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new policy version. The new version automatically becomes current.
func (v *PolicyVersionsAPI) Create(policyID string, content map[string]any, versionName string, baseVersionID string) (*generated.CreatePolicyVersion, error) {
	return v.CreateWithContext(context.Background(), policyID, content, versionName, baseVersionID)
}

// CreateWithContext creates a new policy version with a caller-supplied context.
func (v *PolicyVersionsAPI) CreateWithContext(ctx context.Context, policyID string, content map[string]any, versionName string, baseVersionID string) (*generated.CreatePolicyVersion, error) {
	pid, err := parseUUIDParam("policyID", policyID)
	if err != nil {
		return nil, err
	}
	orgID, err := v.policiesAPI.orgUUID()
	if err != nil {
		return nil, err
	}
	body := generated.CreatePolicyVersionRequest{Content: content}
	if versionName != "" {
		body.VersionName = &versionName
	}
	if baseVersionID != "" {
		bid, err := uuid.Parse(baseVersionID)
		if err != nil {
			return nil, fmt.Errorf("invalid baseVersionID: %w", err)
		}
		body.BaseVersionId = &bid
	}
	resp, err := v.policiesAPI.raw().V1PoliciesVersionsCreateWithResponse(ctx, pid, &generated.V1PoliciesVersionsCreateParams{OrganizationId: orgID}, body)
	if err != nil {
		return nil, err
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}
