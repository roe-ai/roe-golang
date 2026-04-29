package roe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/roe-ai/roe-golang/generated"
)

// readAndCloseBody consumes an *http.Response body returned by the lower-level
// generated.Client (WithBody variants without a typed response wrapper). It
// always closes the body, returning either the read bytes or the read error.
func readAndCloseBody(rsp *http.Response) ([]byte, error) {
	if rsp == nil || rsp.Body == nil {
		return nil, nil
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

func stringifyGeneratedStringField(value any) any {
	if value == nil {
		return nil
	}
	if _, ok := value.(string); ok {
		return value
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(encoded)
}

func normalizeBaseAgentTags(value any) {
	switch v := value.(type) {
	case map[string]any:
		if tags, ok := v["tags"]; ok {
			v["tags"] = stringifyGeneratedStringField(tags)
		}
		for _, child := range v {
			normalizeBaseAgentTags(child)
		}
	case []any:
		for _, child := range v {
			normalizeBaseAgentTags(child)
		}
	}
}

func decodeAgentResponse[T any](httpResp *http.Response, context string) (*T, error) {
	body, rErr := readAndCloseBody(httpResp)
	if rErr != nil {
		return nil, fmt.Errorf("%s: read body: %w", context, rErr)
	}
	if err := errorFromResponse(httpResp, body); err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, fmt.Errorf("%s: empty response body", context)
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("%s: parse response body: %w", context, err)
	}
	normalizeBaseAgentTags(payload)
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: normalize response body: %w", context, err)
	}
	var out T
	if err := json.Unmarshal(normalized, &out); err != nil {
		return nil, fmt.Errorf("%s: parse normalized response body: %w", context, err)
	}
	return &out, nil
}

const maxBatchSize = 1000

// AgentsAPI manages agent operations. All methods delegate to the generated
// raw client; the wrapper layer provides ergonomic argument coercion (string
// IDs → UUID, optional pointers, multipart pre-processing for dynamic inputs)
// and translates non-2xx responses to the typed RoeAPIException family via
// errorFromResponse.
type AgentsAPI struct {
	cfg      Config
	client   *RoeClient
	Versions *AgentVersionsAPI
	Jobs     *AgentJobsAPI
}

func newAgentsAPI(client *RoeClient) *AgentsAPI {
	api := &AgentsAPI{cfg: client.Config, client: client}
	api.Versions = &AgentVersionsAPI{agentsAPI: api}
	api.Jobs = &AgentJobsAPI{agentsAPI: api}
	return api
}

func (a *AgentsAPI) raw() *generated.ClientWithResponses { return a.client.raw }
func (a *AgentsAPI) http() *httpClient                   { return a.client.http }

func (a *AgentsAPI) orgUUID() (*openapi_types.UUID, error) {
	if a.cfg.OrganizationID == "" {
		return nil, nil
	}
	id, err := uuid.Parse(a.cfg.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	return &id, nil
}

// requireOrgUUID returns the organization UUID as a non-pointer for endpoints
// (like create) where it appears in a required request body field.
func (a *AgentsAPI) requireOrgUUID() (openapi_types.UUID, error) {
	if a.cfg.OrganizationID == "" {
		return openapi_types.UUID{}, fmt.Errorf("organization ID is required")
	}
	id, err := uuid.Parse(a.cfg.OrganizationID)
	if err != nil {
		return openapi_types.UUID{}, fmt.Errorf("invalid organization ID: %w", err)
	}
	return id, nil
}

// List returns paginated agents.
func (a *AgentsAPI) List(page, pageSize int) (*generated.PaginatedBaseAgentList, error) {
	return a.ListWithContext(context.Background(), page, pageSize)
}

// ListWithContext returns paginated agents with a caller-supplied context.
func (a *AgentsAPI) ListWithContext(ctx context.Context, page, pageSize int) (*generated.PaginatedBaseAgentList, error) {
	orgID, err := a.requireOrgUUID()
	if err != nil {
		return nil, err
	}
	params := &generated.V1AgentsListParams{OrganizationId: orgID}
	if page > 0 {
		params.Page = &page
	}
	if pageSize > 0 {
		params.PageSize = &pageSize
	}
	httpResp, err := a.raw().V1AgentsList(ctx, params)
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[generated.PaginatedBaseAgentList](httpResp, "list agents")
}

// Retrieve fetches an agent by ID.
func (a *AgentsAPI) Retrieve(agentID string) (*generated.BaseAgent, error) {
	return a.RetrieveWithContext(context.Background(), agentID)
}

// RetrieveWithContext fetches an agent by ID with a caller-supplied context.
func (a *AgentsAPI) RetrieveWithContext(ctx context.Context, agentID string) (*generated.BaseAgent, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return nil, err
	}
	httpResp, err := a.raw().V1AgentsRetrieve(ctx, id, &generated.V1AgentsRetrieveParams{OrganizationId: orgID})
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[generated.BaseAgent](httpResp, "retrieve agent")
}

// Create creates a new agent.
func (a *AgentsAPI) Create(name, engineClassID string, inputDefs []map[string]any, engineConfig map[string]any, versionName, description string) (*generated.BaseAgent, error) {
	return a.CreateWithContext(context.Background(), name, engineClassID, inputDefs, engineConfig, versionName, description)
}

// CreateWithContext creates a new agent with a caller-supplied context.
func (a *AgentsAPI) CreateWithContext(ctx context.Context, name, engineClassID string, inputDefs []map[string]any, engineConfig map[string]any, versionName, description string) (*generated.BaseAgent, error) {
	orgID, err := a.requireOrgUUID()
	if err != nil {
		return nil, err
	}
	body := generated.BaseAgentCreateRequest{
		Name:             name,
		EngineClassId:    engineClassID,
		OrganizationId:   orgID,
		InputDefinitions: inputDefs,
		EngineConfig:     engineConfig,
	}
	if versionName != "" {
		body.VersionName = &versionName
	}
	if description != "" {
		body.Description = &description
	}
	httpResp, err := a.raw().V1AgentsCreate(ctx, &generated.V1AgentsCreateParams{}, body)
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[generated.BaseAgent](httpResp, "create agent")
}

// Update updates an agent's metadata. Pass nil for fields you don't want to change.
// Uses PATCH semantics (only supplied fields are sent).
func (a *AgentsAPI) Update(agentID string, name *string, disableCache, cacheFailedJobs *bool) (*generated.BaseAgent, error) {
	return a.UpdateWithContext(context.Background(), agentID, name, disableCache, cacheFailedJobs)
}

// UpdateWithContext updates an agent with a caller-supplied context.
func (a *AgentsAPI) UpdateWithContext(ctx context.Context, agentID string, name *string, disableCache, cacheFailedJobs *bool) (*generated.BaseAgent, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return nil, err
	}
	body := generated.PatchedBaseAgentUpdateRequest{
		Name:            name,
		DisableCache:    disableCache,
		CacheFailedJobs: cacheFailedJobs,
	}
	httpResp, err := a.raw().V1AgentsPartialUpdate(ctx, id, &generated.V1AgentsPartialUpdateParams{OrganizationId: orgID}, body)
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[generated.BaseAgent](httpResp, "update agent")
}

// Delete removes an agent.
func (a *AgentsAPI) Delete(agentID string) error {
	return a.DeleteWithContext(context.Background(), agentID)
}

// DeleteWithContext removes an agent with a caller-supplied context.
func (a *AgentsAPI) DeleteWithContext(ctx context.Context, agentID string) error {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return err
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return err
	}
	resp, err := a.raw().V1AgentsDestroyWithResponse(ctx, id, &generated.V1AgentsDestroyParams{OrganizationId: orgID})
	if err != nil {
		return err
	}
	return errorFromResponse(resp.HTTPResponse, resp.Body)
}

// Duplicate clones an agent.
func (a *AgentsAPI) Duplicate(agentID string) (*generated.AgentVersion, error) {
	return a.DuplicateWithContext(context.Background(), agentID)
}

// DuplicateWithContext clones an agent with a caller-supplied context.
func (a *AgentsAPI) DuplicateWithContext(ctx context.Context, agentID string) (*generated.AgentVersion, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return nil, err
	}
	httpResp, err := a.raw().V1AgentsDuplicateCreate(ctx, id, &generated.V1AgentsDuplicateCreateParams{OrganizationId: orgID})
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[generated.AgentVersion](httpResp, "duplicate agent")
}

// Run starts an async job for the given agent. Inputs may include FileUpload
// values, file path strings, *bytes.Buffer/Reader, raw bytes, plain strings,
// or scalars; the dynamic-input pre-processor builds the appropriate
// multipart/form or urlencoded body.
func (a *AgentsAPI) Run(agentID string, timeoutSeconds int, inputs map[string]any, metadata map[string]any) (*Job, error) {
	return a.RunWithContext(context.Background(), agentID, timeoutSeconds, inputs, metadata)
}

// RunWithContext starts an async job with a caller-supplied context.
func (a *AgentsAPI) RunWithContext(ctx context.Context, agentID string, timeoutSeconds int, inputs map[string]any, metadata map[string]any) (*Job, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return nil, err
	}
	body, contentType, err := a.http().dynamicInputsRequest(inputs, metadata)
	if err != nil {
		return nil, fmt.Errorf("run agent %s: %w", agentID, err)
	}
	resp, err := a.raw().V1AgentsRunAsyncCreateWithBodyWithResponse(ctx, id, &generated.V1AgentsRunAsyncCreateParams{OrganizationId: orgID}, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("run agent %s: %w", agentID, err)
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	jobID, err := decodeJobID(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("run agent %s: %w", agentID, err)
	}
	return newJob(a, jobID, timeoutSeconds), nil
}

// RunMany submits batch jobs. The payload is JSON-encoded
// {inputs: [...], metadata: ...} per the backend's contract; the generated
// AgentRunAsyncManyRequestRequest type lacks a metadata field, so we route
// through the *WithBody* variant with a hand-marshaled body.
func (a *AgentsAPI) RunMany(agentID string, batchInputs []map[string]any, timeoutSeconds int, metadata map[string]any) (*JobBatch, error) {
	return a.RunManyWithContext(context.Background(), agentID, batchInputs, timeoutSeconds, metadata)
}

// RunManyWithContext submits batch jobs with a caller-supplied context.
func (a *AgentsAPI) RunManyWithContext(ctx context.Context, agentID string, batchInputs []map[string]any, timeoutSeconds int, metadata map[string]any) (*JobBatch, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	if len(batchInputs) == 0 {
		return nil, fmt.Errorf("batchInputs cannot be empty")
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return nil, err
	}
	jobIDs := []string{}
	for _, chunk := range chunkAny(batchInputs, maxBatchSize) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		payload := map[string]any{"inputs": chunk}
		if metadata != nil {
			payload["metadata"] = metadata
		}
		buf, mErr := json.Marshal(payload)
		if mErr != nil {
			return nil, fmt.Errorf("run-many: marshal payload: %w", mErr)
		}
		// Spec declares JSON200 as an object (additionalProperties), but the
		// backend returns a JSON array of UUID strings. Use the lower-level
		// Client method (no typed response wrapper) to avoid a parse error
		// from the spec-mismatched WithResponse path; we read the body and
		// unmarshal as []string ourselves.
		httpResp, err := a.raw().AgentsRunAsyncMany5WithBody(ctx, id, &generated.AgentsRunAsyncMany5Params{OrganizationId: orgID}, "application/json", bytes.NewReader(buf))
		if err != nil {
			return nil, err
		}
		body, rErr := readAndCloseBody(httpResp)
		if rErr != nil {
			return nil, fmt.Errorf("run-many: read body: %w", rErr)
		}
		if err := errorFromResponse(httpResp, body); err != nil {
			return nil, err
		}
		var ids []string
		if uErr := json.Unmarshal(body, &ids); uErr != nil {
			return nil, fmt.Errorf("run-many: parse job IDs: %w", uErr)
		}
		jobIDs = append(jobIDs, ids...)
	}
	return newJobBatch(a, jobIDs, timeoutSeconds), nil
}

// RunSync runs synchronously and returns outputs.
func (a *AgentsAPI) RunSync(agentID string, inputs map[string]any, metadata map[string]any) ([]AgentDatum, error) {
	return a.RunSyncWithContext(context.Background(), agentID, inputs, metadata)
}

// RunSyncWithContext runs synchronously with a caller-supplied context.
func (a *AgentsAPI) RunSyncWithContext(ctx context.Context, agentID string, inputs map[string]any, metadata map[string]any) ([]AgentDatum, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return nil, err
	}
	body, contentType, err := a.http().dynamicInputsRequest(inputs, metadata)
	if err != nil {
		return nil, fmt.Errorf("run agent %s sync: %w", agentID, err)
	}
	resp, err := a.raw().AgentsRun2WithBodyWithResponse(ctx, id, &generated.AgentsRun2Params{OrganizationId: orgID}, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("run agent %s sync: %w", agentID, err)
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	var data []AgentDatum
	if uErr := json.Unmarshal(resp.Body, &data); uErr != nil {
		return nil, fmt.Errorf("run agent %s sync: parse outputs: %w", agentID, uErr)
	}
	return data, nil
}

// RunVersion runs a specific version asynchronously.
func (a *AgentsAPI) RunVersion(agentID, versionID string, timeoutSeconds int, inputs map[string]any, metadata map[string]any) (*Job, error) {
	return a.RunVersionWithContext(context.Background(), agentID, versionID, timeoutSeconds, inputs, metadata)
}

// RunVersionWithContext runs a specific version asynchronously with a caller-supplied context.
func (a *AgentsAPI) RunVersionWithContext(ctx context.Context, agentID, versionID string, timeoutSeconds int, inputs map[string]any, metadata map[string]any) (*Job, error) {
	aID, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	vID, err := parseUUIDParam("versionID", versionID)
	if err != nil {
		return nil, err
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return nil, err
	}
	body, contentType, err := a.http().dynamicInputsRequest(inputs, metadata)
	if err != nil {
		return nil, fmt.Errorf("run agent %s version %s: %w", agentID, versionID, err)
	}
	resp, err := a.raw().V1AgentsRunVersionsAsyncCreateWithBodyWithResponse(ctx, aID, vID, &generated.V1AgentsRunVersionsAsyncCreateParams{OrganizationId: orgID}, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("run agent %s version %s: %w", agentID, versionID, err)
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	jobID, err := decodeJobID(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("run agent %s version %s: %w", agentID, versionID, err)
	}
	return newJob(a, jobID, timeoutSeconds), nil
}

// RunVersionSync runs a specific version synchronously.
func (a *AgentsAPI) RunVersionSync(agentID, versionID string, inputs map[string]any, metadata map[string]any) ([]AgentDatum, error) {
	return a.RunVersionSyncWithContext(context.Background(), agentID, versionID, inputs, metadata)
}

// RunVersionSyncWithContext runs a specific version synchronously with a caller-supplied context.
func (a *AgentsAPI) RunVersionSyncWithContext(ctx context.Context, agentID, versionID string, inputs map[string]any, metadata map[string]any) ([]AgentDatum, error) {
	aID, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	vID, err := parseUUIDParam("versionID", versionID)
	if err != nil {
		return nil, err
	}
	orgID, err := a.orgUUID()
	if err != nil {
		return nil, err
	}
	body, contentType, err := a.http().dynamicInputsRequest(inputs, metadata)
	if err != nil {
		return nil, fmt.Errorf("run agent %s version %s sync: %w", agentID, versionID, err)
	}
	resp, err := a.raw().AgentsRunVersion2WithBodyWithResponse(ctx, aID, vID, &generated.AgentsRunVersion2Params{OrganizationId: orgID}, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("run agent %s version %s sync: %w", agentID, versionID, err)
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	var data []AgentDatum
	if uErr := json.Unmarshal(resp.Body, &data); uErr != nil {
		return nil, fmt.Errorf("run agent %s version %s sync: parse outputs: %w", agentID, versionID, uErr)
	}
	return data, nil
}

// AgentVersionsAPI handles version operations.
type AgentVersionsAPI struct {
	agentsAPI *AgentsAPI
}

func (v *AgentVersionsAPI) raw() *generated.ClientWithResponses { return v.agentsAPI.raw() }

// List returns all versions of an agent.
func (v *AgentVersionsAPI) List(agentID string) (*[]generated.AgentVersion, error) {
	return v.ListWithContext(context.Background(), agentID)
}

// ListWithContext returns all versions of an agent with a caller-supplied context.
func (v *AgentVersionsAPI) ListWithContext(ctx context.Context, agentID string) (*[]generated.AgentVersion, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	orgID, err := v.agentsAPI.orgUUID()
	if err != nil {
		return nil, err
	}
	httpResp, err := v.raw().V1AgentsVersionsList(ctx, id, &generated.V1AgentsVersionsListParams{OrganizationId: orgID})
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[[]generated.AgentVersion](httpResp, "list agent versions")
}

// Retrieve fetches a specific agent version.
func (v *AgentVersionsAPI) Retrieve(agentID, versionID string, getSupportsEval *bool) (*generated.AgentVersion, error) {
	return v.RetrieveWithContext(context.Background(), agentID, versionID, getSupportsEval)
}

// RetrieveWithContext fetches a specific agent version with a caller-supplied context.
func (v *AgentVersionsAPI) RetrieveWithContext(ctx context.Context, agentID, versionID string, getSupportsEval *bool) (*generated.AgentVersion, error) {
	aID, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	vID, err := parseUUIDParam("versionID", versionID)
	if err != nil {
		return nil, err
	}
	orgID, err := v.agentsAPI.orgUUID()
	if err != nil {
		return nil, err
	}
	params := &generated.V1AgentsVersionsRetrieveParams{
		OrganizationId:  orgID,
		GetSupportsEval: getSupportsEval,
	}
	httpResp, err := v.raw().V1AgentsVersionsRetrieve(ctx, aID, vID, params)
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[generated.AgentVersion](httpResp, "retrieve agent version")
}

// RetrieveCurrent fetches the agent's current version.
func (v *AgentVersionsAPI) RetrieveCurrent(agentID string) (*generated.AgentVersion, error) {
	return v.RetrieveCurrentWithContext(context.Background(), agentID)
}

// RetrieveCurrentWithContext fetches the current version with a caller-supplied context.
func (v *AgentVersionsAPI) RetrieveCurrentWithContext(ctx context.Context, agentID string) (*generated.AgentVersion, error) {
	return v.RetrieveCurrentWithEvalWithContext(ctx, agentID, nil)
}

// RetrieveCurrentWithEval fetches the current version with optional eval flag.
func (v *AgentVersionsAPI) RetrieveCurrentWithEval(agentID string, getSupportsEval *bool) (*generated.AgentVersion, error) {
	return v.RetrieveCurrentWithEvalWithContext(context.Background(), agentID, getSupportsEval)
}

// RetrieveCurrentWithEvalWithContext fetches the current version with caller-supplied context.
func (v *AgentVersionsAPI) RetrieveCurrentWithEvalWithContext(ctx context.Context, agentID string, getSupportsEval *bool) (*generated.AgentVersion, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	orgID, err := v.agentsAPI.orgUUID()
	if err != nil {
		return nil, err
	}
	params := &generated.V1AgentsVersionsCurrentRetrieveParams{
		OrganizationId:  orgID,
		GetSupportsEval: getSupportsEval,
	}
	httpResp, err := v.raw().V1AgentsVersionsCurrentRetrieve(ctx, id, params)
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[generated.AgentVersion](httpResp, "retrieve current agent version")
}

// Create creates a new agent version.
func (v *AgentVersionsAPI) Create(agentID string, inputDefs []map[string]any, engineConfig map[string]any, versionName, description string) (*generated.AgentVersion, error) {
	return v.CreateWithContext(context.Background(), agentID, inputDefs, engineConfig, versionName, description)
}

// CreateWithContext creates a new agent version with a caller-supplied context.
func (v *AgentVersionsAPI) CreateWithContext(ctx context.Context, agentID string, inputDefs []map[string]any, engineConfig map[string]any, versionName, description string) (*generated.AgentVersion, error) {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return nil, err
	}
	orgID, err := v.agentsAPI.orgUUID()
	if err != nil {
		return nil, err
	}
	body := generated.AgentVersionCreateRequest{
		InputDefinitions: inputDefs,
		EngineConfig:     engineConfig,
	}
	if versionName != "" {
		body.VersionName = &versionName
	}
	if description != "" {
		body.Description = &description
	}
	httpResp, err := v.raw().V1AgentsVersionsCreate(ctx, id, &generated.V1AgentsVersionsCreateParams{OrganizationId: orgID}, body)
	if err != nil {
		return nil, err
	}
	return decodeAgentResponse[generated.AgentVersion](httpResp, "create agent version")
}

// Update edits an agent version's metadata. Pass empty strings to leave fields unchanged.
// Uses PATCH semantics.
func (v *AgentVersionsAPI) Update(agentID, versionID, versionName, description string) error {
	return v.UpdateWithContext(context.Background(), agentID, versionID, versionName, description)
}

// UpdateWithContext edits an agent version with a caller-supplied context.
func (v *AgentVersionsAPI) UpdateWithContext(ctx context.Context, agentID, versionID, versionName, description string) error {
	aID, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return err
	}
	vID, err := parseUUIDParam("versionID", versionID)
	if err != nil {
		return err
	}
	orgID, err := v.agentsAPI.orgUUID()
	if err != nil {
		return err
	}
	body := generated.PatchedPatchedAgentVersionUpdateRequestRequest{}
	if versionName != "" {
		body.VersionName = &versionName
	}
	if description != "" {
		body.Description = &description
	}
	httpResp, err := v.raw().V1AgentsVersionsPartialUpdate(ctx, aID, vID, &generated.V1AgentsVersionsPartialUpdateParams{OrganizationId: orgID}, body)
	if err != nil {
		return err
	}
	bodyBytes, rErr := readAndCloseBody(httpResp)
	if rErr != nil {
		return fmt.Errorf("update agent version: read body: %w", rErr)
	}
	return errorFromResponse(httpResp, bodyBytes)
}

// Delete removes an agent version.
func (v *AgentVersionsAPI) Delete(agentID, versionID string) error {
	return v.DeleteWithContext(context.Background(), agentID, versionID)
}

// DeleteWithContext removes an agent version with a caller-supplied context.
func (v *AgentVersionsAPI) DeleteWithContext(ctx context.Context, agentID, versionID string) error {
	aID, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return err
	}
	vID, err := parseUUIDParam("versionID", versionID)
	if err != nil {
		return err
	}
	orgID, err := v.agentsAPI.orgUUID()
	if err != nil {
		return err
	}
	resp, err := v.raw().V1AgentsVersionsDestroyWithResponse(ctx, aID, vID, &generated.V1AgentsVersionsDestroyParams{OrganizationId: orgID})
	if err != nil {
		return err
	}
	return errorFromResponse(resp.HTTPResponse, resp.Body)
}

// AgentJobsAPI handles job operations. Single-job endpoints return generated
// types' content as the existing handwritten Job-ergonomic structs (see
// types.go) so Job.WaitContext / JobBatch.WaitContext keep their internal
// logic. Bulk endpoints' wire format includes per-item job IDs, which the
// generated request/response schemas omit — those endpoints are routed
// through *WithBody* variants and parsed into AgentJobStatusBatch /
// AgentJobResultBatch by hand.
type AgentJobsAPI struct {
	agentsAPI *AgentsAPI
}

func (j *AgentJobsAPI) raw() *generated.ClientWithResponses { return j.agentsAPI.raw() }

// RetrieveStatus fetches a job's current status.
func (j *AgentJobsAPI) RetrieveStatus(jobID string) (AgentJobStatus, error) {
	return j.RetrieveStatusWithContext(context.Background(), jobID)
}

// RetrieveStatusWithContext fetches a job's current status with a caller-supplied context.
func (j *AgentJobsAPI) RetrieveStatusWithContext(ctx context.Context, jobID string) (AgentJobStatus, error) {
	id, err := parseUUIDParam("jobID", jobID)
	if err != nil {
		return AgentJobStatus{}, err
	}
	orgID, err := j.agentsAPI.orgUUID()
	if err != nil {
		return AgentJobStatus{}, err
	}
	httpResp, err := j.raw().V1AgentsJobsStatusRetrieve(ctx, id, &generated.V1AgentsJobsStatusRetrieveParams{OrganizationId: orgID})
	if err != nil {
		return AgentJobStatus{}, err
	}
	body, rErr := readAndCloseBody(httpResp)
	if rErr != nil {
		return AgentJobStatus{}, fmt.Errorf("retrieve status: read body: %w", rErr)
	}
	if err := errorFromResponse(httpResp, body); err != nil {
		return AgentJobStatus{}, err
	}
	// Parse the wire body into the handwritten polling-friendly struct so
	// JobStatus is the typed enum and the timestamp is float64.
	var status AgentJobStatus
	if uErr := json.Unmarshal(body, &status); uErr != nil {
		return AgentJobStatus{}, fmt.Errorf("retrieve status: parse: %w", uErr)
	}
	return status, nil
}

// RetrieveResult fetches a job's result.
func (j *AgentJobsAPI) RetrieveResult(jobID string) (AgentJobResult, error) {
	return j.RetrieveResultWithContext(context.Background(), jobID)
}

// RetrieveResultWithContext fetches a job's result with a caller-supplied context.
func (j *AgentJobsAPI) RetrieveResultWithContext(ctx context.Context, jobID string) (AgentJobResult, error) {
	id, err := parseUUIDParam("jobID", jobID)
	if err != nil {
		return AgentJobResult{}, err
	}
	orgID, err := j.agentsAPI.orgUUID()
	if err != nil {
		return AgentJobResult{}, err
	}
	httpResp, err := j.raw().V1AgentsJobsResultRetrieve(ctx, id, &generated.V1AgentsJobsResultRetrieveParams{OrganizationId: orgID})
	if err != nil {
		return AgentJobResult{}, err
	}
	body, rErr := readAndCloseBody(httpResp)
	if rErr != nil {
		return AgentJobResult{}, fmt.Errorf("retrieve result: read body: %w", rErr)
	}
	if err := errorFromResponse(httpResp, body); err != nil {
		return AgentJobResult{}, err
	}
	var result AgentJobResult
	if uErr := json.Unmarshal(body, &result); uErr != nil {
		return AgentJobResult{}, fmt.Errorf("retrieve result: parse: %w", uErr)
	}
	return result, nil
}

// RetrieveStatusMany fetches statuses for multiple jobs.
func (j *AgentJobsAPI) RetrieveStatusMany(jobIDs []string) ([]AgentJobStatusBatch, error) {
	return j.RetrieveStatusManyWithContext(context.Background(), jobIDs)
}

// RetrieveStatusManyWithContext fetches statuses with a caller-supplied context.
func (j *AgentJobsAPI) RetrieveStatusManyWithContext(ctx context.Context, jobIDs []string) ([]AgentJobStatusBatch, error) {
	if len(jobIDs) == 0 {
		return nil, nil
	}
	orgID, err := j.agentsAPI.orgUUID()
	if err != nil {
		return nil, err
	}

	order := make(map[string]int, len(jobIDs))
	for idx, id := range jobIDs {
		order[id] = idx
	}
	results := make([]AgentJobStatusBatch, len(jobIDs))
	for i, id := range jobIDs {
		results[i] = AgentJobStatusBatch{ID: id}
	}
	received := make(map[string]bool, len(jobIDs))

	for _, chunk := range chunkStrings(jobIDs, maxBatchSize) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// Wire format includes a per-item `id` field; the generated AgentJobStatus
		// schema omits it. Use the lower-level Client (no typed response wrapper)
		// and parse the body into the handwritten Batch type ourselves — the
		// WithResponse parser would still succeed for this endpoint, but we keep
		// the bulk endpoints uniform by going through the same code path.
		buf, mErr := json.Marshal(map[string]any{"job_ids": chunk})
		if mErr != nil {
			return nil, fmt.Errorf("retrieve statuses: marshal: %w", mErr)
		}
		httpResp, err := j.raw().V1AgentsJobsStatusesCreateWithBody(ctx, &generated.V1AgentsJobsStatusesCreateParams{OrganizationId: orgID}, "application/json", bytes.NewReader(buf))
		if err != nil {
			return nil, fmt.Errorf("retrieve job statuses: %w", err)
		}
		body, rErr := readAndCloseBody(httpResp)
		if rErr != nil {
			return nil, fmt.Errorf("retrieve job statuses: read body: %w", rErr)
		}
		if err := errorFromResponse(httpResp, body); err != nil {
			return nil, fmt.Errorf("retrieve job statuses: %w", err)
		}
		var batch []AgentJobStatusBatch
		if uErr := json.Unmarshal(body, &batch); uErr != nil {
			return nil, fmt.Errorf("retrieve job statuses: parse: %w", uErr)
		}
		for _, st := range batch {
			if idx, ok := order[st.ID]; ok {
				results[idx] = st
				received[st.ID] = true
			}
		}
	}

	var missing []string
	for _, id := range jobIDs {
		if !received[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("jobs not found in status response: %v", missing)
	}
	return results, nil
}

// RetrieveResultMany fetches results for multiple jobs.
func (j *AgentJobsAPI) RetrieveResultMany(jobIDs []string) ([]AgentJobResultBatch, error) {
	return j.RetrieveResultManyWithContext(context.Background(), jobIDs)
}

// RetrieveResultManyWithContext fetches results with a caller-supplied context.
func (j *AgentJobsAPI) RetrieveResultManyWithContext(ctx context.Context, jobIDs []string) ([]AgentJobResultBatch, error) {
	if len(jobIDs) == 0 {
		return nil, nil
	}
	orgID, err := j.agentsAPI.orgUUID()
	if err != nil {
		return nil, err
	}

	order := make(map[string]int, len(jobIDs))
	for idx, id := range jobIDs {
		order[id] = idx
	}
	results := make([]AgentJobResultBatch, len(jobIDs))
	for i, id := range jobIDs {
		results[i] = AgentJobResultBatch{ID: id}
	}
	received := make(map[string]bool, len(jobIDs))

	for _, chunk := range chunkStrings(jobIDs, maxBatchSize) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		buf, mErr := json.Marshal(map[string]any{"job_ids": chunk})
		if mErr != nil {
			return nil, fmt.Errorf("retrieve results: marshal: %w", mErr)
		}
		// Wire format is a JSON array of items with per-item id; the spec
		// declares a paginated response, so the WithResponse parser would
		// reject the body. Use the lower-level Client and parse manually.
		httpResp, err := j.raw().V1AgentsJobsResultsCreateWithBody(ctx, &generated.V1AgentsJobsResultsCreateParams{OrganizationId: orgID}, "application/json", bytes.NewReader(buf))
		if err != nil {
			return nil, fmt.Errorf("retrieve job results: %w", err)
		}
		body, rErr := readAndCloseBody(httpResp)
		if rErr != nil {
			return nil, fmt.Errorf("retrieve job results: read body: %w", rErr)
		}
		if err := errorFromResponse(httpResp, body); err != nil {
			return nil, fmt.Errorf("retrieve job results: %w", err)
		}
		var batch []AgentJobResultBatch
		if uErr := json.Unmarshal(body, &batch); uErr != nil {
			return nil, fmt.Errorf("retrieve job results: parse: %w", uErr)
		}
		for _, st := range batch {
			if idx, ok := order[st.ID]; ok {
				results[idx] = st
				received[st.ID] = true
			}
		}
	}

	var missing []string
	for _, id := range jobIDs {
		if !received[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("jobs not found in results response: %v", missing)
	}
	return results, nil
}

// DownloadReference fetches a reference resource as raw bytes.
func (j *AgentJobsAPI) DownloadReference(jobID, resourceID string, asAttachment bool) ([]byte, error) {
	return j.DownloadReferenceWithContext(context.Background(), jobID, resourceID, asAttachment)
}

// DownloadReferenceWithContext fetches a reference with a caller-supplied context.
func (j *AgentJobsAPI) DownloadReferenceWithContext(ctx context.Context, jobID, resourceID string, asAttachment bool) ([]byte, error) {
	id, err := parseUUIDParam("jobID", jobID)
	if err != nil {
		return nil, err
	}
	if _, err := parseUUIDParam("resourceID", resourceID); err != nil {
		return nil, err
	}
	orgID, err := j.agentsAPI.orgUUID()
	if err != nil {
		return nil, err
	}
	params := &generated.V1AgentsJobsReferencesRetrieveParams{OrganizationId: orgID}
	// The OpenAPI spec doesn't declare a `download` query parameter, but the
	// backend honors it when set to "true" to force the proxy to attach the
	// file rather than render inline. Inject via a request editor.
	editors := []generated.RequestEditorFn{}
	if asAttachment {
		editors = append(editors, func(ctx context.Context, req *http.Request) error {
			q := req.URL.Query()
			q.Set("download", "true")
			req.URL.RawQuery = q.Encode()
			return nil
		})
	}
	httpResp, err := j.raw().V1AgentsJobsReferencesRetrieve(ctx, id, resourceID, params, editors...)
	if err != nil {
		return nil, err
	}
	body, rErr := readAndCloseBody(httpResp)
	if rErr != nil {
		return nil, fmt.Errorf("download reference: read body: %w", rErr)
	}
	if err := errorFromResponse(httpResp, body); err != nil {
		return nil, err
	}
	return body, nil
}

// DeleteData purges blob data for a job.
func (j *AgentJobsAPI) DeleteData(jobID string) (JobDataDeleteResponse, error) {
	return j.DeleteDataWithContext(context.Background(), jobID)
}

// DeleteDataWithContext purges blob data with a caller-supplied context.
func (j *AgentJobsAPI) DeleteDataWithContext(ctx context.Context, jobID string) (JobDataDeleteResponse, error) {
	id, err := parseUUIDParam("jobID", jobID)
	if err != nil {
		return JobDataDeleteResponse{}, err
	}
	orgID, err := j.agentsAPI.orgUUID()
	if err != nil {
		return JobDataDeleteResponse{}, err
	}
	httpResp, err := j.raw().V1AgentsJobsDeleteDataCreate(ctx, id, &generated.V1AgentsJobsDeleteDataCreateParams{OrganizationId: orgID})
	if err != nil {
		return JobDataDeleteResponse{}, err
	}
	body, rErr := readAndCloseBody(httpResp)
	if rErr != nil {
		return JobDataDeleteResponse{}, fmt.Errorf("delete data: read body: %w", rErr)
	}
	if err := errorFromResponse(httpResp, body); err != nil {
		return JobDataDeleteResponse{}, err
	}
	var out JobDataDeleteResponse
	if uErr := json.Unmarshal(body, &out); uErr != nil {
		return JobDataDeleteResponse{}, fmt.Errorf("delete data: parse: %w", uErr)
	}
	return out, nil
}

// Cancel cancels a running job.
func (j *AgentJobsAPI) Cancel(jobID string) error {
	return j.CancelWithContext(context.Background(), jobID)
}

// CancelWithContext cancels a running job with a caller-supplied context.
func (j *AgentJobsAPI) CancelWithContext(ctx context.Context, jobID string) error {
	id, err := parseUUIDParam("jobID", jobID)
	if err != nil {
		return err
	}
	orgID, err := j.agentsAPI.orgUUID()
	if err != nil {
		return err
	}
	httpResp, err := j.raw().V1AgentsJobsCancelCreate(ctx, id, &generated.V1AgentsJobsCancelCreateParams{OrganizationId: orgID})
	if err != nil {
		return err
	}
	body, rErr := readAndCloseBody(httpResp)
	if rErr != nil {
		return fmt.Errorf("cancel job: read body: %w", rErr)
	}
	return errorFromResponse(httpResp, body)
}

// CancelAll cancels all running jobs for an agent.
func (j *AgentJobsAPI) CancelAll(agentID string) error {
	return j.CancelAllWithContext(context.Background(), agentID)
}

// CancelAllWithContext cancels all running jobs for an agent with a caller-supplied context.
func (j *AgentJobsAPI) CancelAllWithContext(ctx context.Context, agentID string) error {
	id, err := parseUUIDParam("agentID", agentID)
	if err != nil {
		return err
	}
	orgID, err := j.agentsAPI.orgUUID()
	if err != nil {
		return err
	}
	httpResp, err := j.raw().V1AgentsJobsCancelAllCreate(ctx, id, &generated.V1AgentsJobsCancelAllCreateParams{OrganizationId: orgID})
	if err != nil {
		return err
	}
	body, rErr := readAndCloseBody(httpResp)
	if rErr != nil {
		return fmt.Errorf("cancel all jobs: read body: %w", rErr)
	}
	return errorFromResponse(httpResp, body)
}

// helpers

// decodeJobID extracts an async job ID from a wire response. The async-create
// endpoints return a plain JSON string ("uuid"); this helper accepts a few
// tolerant shapes (raw string, {"job_id":"..."}, or nested) without erroring
// on extraneous JSON whitespace.
func decodeJobID(body []byte) (string, error) {
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		// Some servers may send a bare token without quotes; fall back to a
		// trimmed string only if it parses as a UUID. If both shapes fail,
		// surface both diagnostics — the JSON error alone is misleading
		// because the body may be a partial HTML error page or similar.
		s := bytes.TrimSpace(body)
		if _, uuidErr := uuid.Parse(string(s)); uuidErr != nil {
			return "", fmt.Errorf("parse job_id: not JSON (%v); not UUID (%w)", err, uuidErr)
		}
		return string(s), nil
	}
	switch v := raw.(type) {
	case string:
		return v, nil
	case map[string]any:
		if id, ok := v["job_id"].(string); ok {
			return id, nil
		}
		if id, ok := v["id"].(string); ok {
			return id, nil
		}
	}
	return "", fmt.Errorf("unexpected job_id payload: %s", string(body))
}

func chunkStrings(items []string, size int) [][]string {
	if size <= 0 || size >= len(items) {
		return [][]string{items}
	}
	var chunks [][]string
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

func chunkAny[T any](items []T, size int) [][]T {
	if size <= 0 || size >= len(items) {
		return [][]T{items}
	}
	var chunks [][]T
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}
