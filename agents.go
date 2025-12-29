package roe

import (
	"context"
	"fmt"
)

const maxBatchSize = 1000

// AgentsAPI manages agent operations.
type AgentsAPI struct {
	cfg        Config
	httpClient *httpClient
	Versions   *AgentVersionsAPI
	Jobs       *AgentJobsAPI
}

func newAgentsAPI(cfg Config, httpClient *httpClient) *AgentsAPI {
	api := &AgentsAPI{cfg: cfg, httpClient: httpClient}
	api.Versions = &AgentVersionsAPI{agentsAPI: api}
	api.Jobs = &AgentJobsAPI{agentsAPI: api}
	return api
}

// List returns paginated agents.
func (a *AgentsAPI) List(page, pageSize int) (PaginatedResponse[BaseAgent], error) {
	return a.ListWithContext(context.Background(), page, pageSize)
}

// ListWithContext returns paginated agents with a caller-supplied context.
func (a *AgentsAPI) ListWithContext(ctx context.Context, page, pageSize int) (PaginatedResponse[BaseAgent], error) {
	params := map[string]string{
		"organization_id": a.cfg.OrganizationID,
	}
	if page > 0 {
		params["page"] = fmt.Sprintf("%d", page)
	}
	if pageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", pageSize)
	}
	var resp PaginatedResponse[BaseAgent]
	if err := a.httpClient.getWithContext(ctx, "/v1/agents/", params, &resp); err != nil {
		return PaginatedResponse[BaseAgent]{}, err
	}
	for i := range resp.Results {
		resp.Results[i].setAgentsAPI(a)
	}
	return resp, nil
}

// Retrieve fetches an agent.
func (a *AgentsAPI) Retrieve(agentID string) (BaseAgent, error) {
	return a.RetrieveWithContext(context.Background(), agentID)
}

// RetrieveWithContext fetches an agent with a caller-supplied context.
func (a *AgentsAPI) RetrieveWithContext(ctx context.Context, agentID string) (BaseAgent, error) {
	if agentID == "" {
		return BaseAgent{}, fmt.Errorf("agentID cannot be empty")
	}
	var resp BaseAgent
	if err := a.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/agents/%s/", agentID), nil, &resp); err != nil {
		return BaseAgent{}, fmt.Errorf("retrieve agent %s: %w", agentID, err)
	}
	resp.setAgentsAPI(a)
	return resp, nil
}

// Create creates a new agent.
func (a *AgentsAPI) Create(name, engineClassID string, inputDefs []map[string]any, engineConfig map[string]any, versionName, description string) (BaseAgent, error) {
	return a.CreateWithContext(context.Background(), name, engineClassID, inputDefs, engineConfig, versionName, description)
}

// CreateWithContext creates a new agent with a caller-supplied context.
func (a *AgentsAPI) CreateWithContext(ctx context.Context, name, engineClassID string, inputDefs []map[string]any, engineConfig map[string]any, versionName, description string) (BaseAgent, error) {
	payload := map[string]any{
		"name":              name,
		"engine_class_id":   engineClassID,
		"organization_id":   a.cfg.OrganizationID,
		"input_definitions": inputDefs,
		"engine_config":     engineConfig,
	}
	if versionName != "" {
		payload["version_name"] = versionName
	}
	if description != "" {
		payload["description"] = description
	}
	var resp BaseAgent
	if err := a.httpClient.postJSONWithContext(ctx, "/v1/agents/", payload, nil, &resp); err != nil {
		return BaseAgent{}, err
	}
	resp.setAgentsAPI(a)
	return resp, nil
}

// Update updates an agent.
func (a *AgentsAPI) Update(agentID string, name string, disableCache, cacheFailedJobs *bool) (BaseAgent, error) {
	return a.UpdateWithContext(context.Background(), agentID, name, disableCache, cacheFailedJobs)
}

// UpdateWithContext updates an agent with a caller-supplied context.
func (a *AgentsAPI) UpdateWithContext(ctx context.Context, agentID string, name string, disableCache, cacheFailedJobs *bool) (BaseAgent, error) {
	payload := map[string]any{}
	if name != "" {
		payload["name"] = name
	}
	if disableCache != nil {
		payload["disable_cache"] = *disableCache
	}
	if cacheFailedJobs != nil {
		payload["cache_failed_jobs"] = *cacheFailedJobs
	}
	var resp BaseAgent
	if err := a.httpClient.putJSONWithContext(ctx, fmt.Sprintf("/v1/agents/%s/", agentID), payload, nil, &resp); err != nil {
		return BaseAgent{}, err
	}
	resp.setAgentsAPI(a)
	return resp, nil
}

// Delete removes an agent.
func (a *AgentsAPI) Delete(agentID string) error {
	return a.DeleteWithContext(context.Background(), agentID)
}

// DeleteWithContext removes an agent with a caller-supplied context.
func (a *AgentsAPI) DeleteWithContext(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agentID cannot be empty")
	}
	if err := a.httpClient.deleteWithContext(ctx, fmt.Sprintf("/v1/agents/%s/", agentID), nil); err != nil {
		return fmt.Errorf("delete agent %s: %w", agentID, err)
	}
	return nil
}

// Duplicate clones an agent.
func (a *AgentsAPI) Duplicate(agentID string) (BaseAgent, error) {
	return a.DuplicateWithContext(context.Background(), agentID)
}

// DuplicateWithContext clones an agent with a caller-supplied context.
func (a *AgentsAPI) DuplicateWithContext(ctx context.Context, agentID string) (BaseAgent, error) {
	var resp struct {
		BaseAgent BaseAgent `json:"base_agent"`
	}
	if err := a.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/agents/%s/duplicate/", agentID), nil, nil, &resp); err != nil {
		return BaseAgent{}, err
	}
	resp.BaseAgent.setAgentsAPI(a)
	return resp.BaseAgent, nil
}

// Run starts an async job for the given agent or version id.
func (a *AgentsAPI) Run(agentID string, timeoutSeconds int, inputs map[string]any) (*Job, error) {
	return a.RunWithContext(context.Background(), agentID, timeoutSeconds, inputs)
}

// RunWithContext starts an async job with a caller-supplied context.
func (a *AgentsAPI) RunWithContext(ctx context.Context, agentID string, timeoutSeconds int, inputs map[string]any) (*Job, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agentID cannot be empty")
	}
	var jobID string
	if err := a.httpClient.postDynamicInputsWithContext(ctx, fmt.Sprintf("/v1/agents/run/%s/async/", agentID), inputs, nil, &jobID); err != nil {
		return nil, fmt.Errorf("run agent %s: %w", agentID, err)
	}
	return newJob(a, jobID, timeoutSeconds), nil
}

// RunMany submits batch jobs.
func (a *AgentsAPI) RunMany(agentID string, batchInputs []map[string]any, timeoutSeconds int) (*JobBatch, error) {
	return a.RunManyWithContext(context.Background(), agentID, batchInputs, timeoutSeconds)
}

// RunManyWithContext submits batch jobs with a caller-supplied context.
func (a *AgentsAPI) RunManyWithContext(ctx context.Context, agentID string, batchInputs []map[string]any, timeoutSeconds int) (*JobBatch, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agentID cannot be empty")
	}
	if len(batchInputs) == 0 {
		return nil, fmt.Errorf("batchInputs cannot be empty")
	}
	jobIDs := []string{}
	for _, chunk := range chunkAny(batchInputs, maxBatchSize) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var ids []string
		payload := map[string]any{"inputs": chunk}
		if err := a.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/agents/run/%s/async/many/", agentID), payload, nil, &ids); err != nil {
			return nil, err
		}
		jobIDs = append(jobIDs, ids...)
	}
	return newJobBatch(a, jobIDs, timeoutSeconds), nil
}

// RunSync runs synchronously and returns outputs.
func (a *AgentsAPI) RunSync(agentID string, inputs map[string]any) ([]AgentDatum, error) {
	return a.RunSyncWithContext(context.Background(), agentID, inputs)
}

// RunSyncWithContext runs synchronously with a caller-supplied context.
func (a *AgentsAPI) RunSyncWithContext(ctx context.Context, agentID string, inputs map[string]any) ([]AgentDatum, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agentID cannot be empty")
	}
	var resp []AgentDatum
	if err := a.httpClient.postDynamicInputsWithContext(ctx, fmt.Sprintf("/v1/agents/run/%s/", agentID), inputs, nil, &resp); err != nil {
		return nil, fmt.Errorf("run agent %s sync: %w", agentID, err)
	}
	return resp, nil
}

// RunVersion runs a specific version asynchronously.
func (a *AgentsAPI) RunVersion(agentID, versionID string, timeoutSeconds int, inputs map[string]any) (*Job, error) {
	return a.RunVersionWithContext(context.Background(), agentID, versionID, timeoutSeconds, inputs)
}

// RunVersionWithContext runs a specific version asynchronously with a caller-supplied context.
func (a *AgentsAPI) RunVersionWithContext(ctx context.Context, agentID, versionID string, timeoutSeconds int, inputs map[string]any) (*Job, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agentID cannot be empty")
	}
	if versionID == "" {
		return nil, fmt.Errorf("versionID cannot be empty")
	}
	var jobID string
	url := fmt.Sprintf("/v1/agents/run/%s/versions/%s/async/", agentID, versionID)
	if err := a.httpClient.postDynamicInputsWithContext(ctx, url, inputs, nil, &jobID); err != nil {
		return nil, fmt.Errorf("run agent %s version %s: %w", agentID, versionID, err)
	}
	return newJob(a, jobID, timeoutSeconds), nil
}

// RunVersionSync runs a specific version synchronously.
func (a *AgentsAPI) RunVersionSync(agentID, versionID string, inputs map[string]any) ([]AgentDatum, error) {
	return a.RunVersionSyncWithContext(context.Background(), agentID, versionID, inputs)
}

// RunVersionSyncWithContext runs a specific version synchronously with a caller-supplied context.
func (a *AgentsAPI) RunVersionSyncWithContext(ctx context.Context, agentID, versionID string, inputs map[string]any) ([]AgentDatum, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agentID cannot be empty")
	}
	if versionID == "" {
		return nil, fmt.Errorf("versionID cannot be empty")
	}
	var resp []AgentDatum
	url := fmt.Sprintf("/v1/agents/run/%s/versions/%s/", agentID, versionID)
	if err := a.httpClient.postDynamicInputsWithContext(ctx, url, inputs, nil, &resp); err != nil {
		return nil, fmt.Errorf("run agent %s version %s sync: %w", agentID, versionID, err)
	}
	return resp, nil
}

// AgentVersionsAPI handles version operations.
type AgentVersionsAPI struct {
	agentsAPI *AgentsAPI
}

type ListVersionsParams struct {
	Page            int
	PageSize        int
	GetSupportsEval *bool
}

func (v *AgentVersionsAPI) List(agentID string) ([]AgentVersion, error) {
	return v.ListWithContext(context.Background(), agentID)
}

func (v *AgentVersionsAPI) ListWithContext(ctx context.Context, agentID string) ([]AgentVersion, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agentID cannot be empty")
	}
	// Note: The versions endpoint returns a raw array, not a paginated response
	var versions []AgentVersion
	if err := v.agentsAPI.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/agents/%s/versions/", agentID), nil, &versions); err != nil {
		return nil, fmt.Errorf("list agent versions: %w", err)
	}
	for i := range versions {
		versions[i].setAgentsAPI(v.agentsAPI)
	}
	return versions, nil
}

func (v *AgentVersionsAPI) ListPaginated(agentID string, params *ListVersionsParams) (PaginatedResponse[AgentVersion], error) {
	return v.ListPaginatedWithContext(context.Background(), agentID, params)
}

func (v *AgentVersionsAPI) ListPaginatedWithContext(ctx context.Context, agentID string, params *ListVersionsParams) (PaginatedResponse[AgentVersion], error) {
	query := map[string]string{}
	if params != nil {
		if params.Page > 0 {
			query["page"] = fmt.Sprintf("%d", params.Page)
		}
		if params.PageSize > 0 {
			query["page_size"] = fmt.Sprintf("%d", params.PageSize)
		}
		if params.GetSupportsEval != nil {
			query["get_supports_eval"] = fmt.Sprintf("%t", *params.GetSupportsEval)
		}
	}
	var resp PaginatedResponse[AgentVersion]
	if err := v.agentsAPI.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/agents/%s/versions/", agentID), query, &resp); err != nil {
		return PaginatedResponse[AgentVersion]{}, err
	}
	for i := range resp.Results {
		resp.Results[i].setAgentsAPI(v.agentsAPI)
	}
	return resp, nil
}

func (v *AgentVersionsAPI) Retrieve(agentID, versionID string, getSupportsEval *bool) (AgentVersion, error) {
	return v.RetrieveWithContext(context.Background(), agentID, versionID, getSupportsEval)
}

func (v *AgentVersionsAPI) RetrieveWithContext(ctx context.Context, agentID, versionID string, getSupportsEval *bool) (AgentVersion, error) {
	params := map[string]string{}
	if getSupportsEval != nil {
		params["get_supports_eval"] = fmt.Sprintf("%t", *getSupportsEval)
	}
	var resp AgentVersion
	if err := v.agentsAPI.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/agents/%s/versions/%s/", agentID, versionID), params, &resp); err != nil {
		return AgentVersion{}, err
	}
	resp.setAgentsAPI(v.agentsAPI)
	return resp, nil
}

func (v *AgentVersionsAPI) RetrieveCurrent(agentID string) (AgentVersion, error) {
	return v.RetrieveCurrentWithContext(context.Background(), agentID)
}

func (v *AgentVersionsAPI) RetrieveCurrentWithContext(ctx context.Context, agentID string) (AgentVersion, error) {
	return v.RetrieveCurrentWithEvalWithContext(ctx, agentID, nil)
}

func (v *AgentVersionsAPI) RetrieveCurrentWithEval(agentID string, getSupportsEval *bool) (AgentVersion, error) {
	return v.RetrieveCurrentWithEvalWithContext(context.Background(), agentID, getSupportsEval)
}

func (v *AgentVersionsAPI) RetrieveCurrentWithEvalWithContext(ctx context.Context, agentID string, getSupportsEval *bool) (AgentVersion, error) {
	params := map[string]string{}
	if getSupportsEval != nil {
		params["get_supports_eval"] = fmt.Sprintf("%t", *getSupportsEval)
	}
	var resp AgentVersion
	if err := v.agentsAPI.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/agents/%s/versions/current/", agentID), params, &resp); err != nil {
		return AgentVersion{}, err
	}
	resp.setAgentsAPI(v.agentsAPI)
	return resp, nil
}

func (v *AgentVersionsAPI) Create(agentID string, inputDefs []map[string]any, engineConfig map[string]any, versionName, description string) (AgentVersion, error) {
	return v.CreateWithContext(context.Background(), agentID, inputDefs, engineConfig, versionName, description)
}

func (v *AgentVersionsAPI) CreateWithContext(ctx context.Context, agentID string, inputDefs []map[string]any, engineConfig map[string]any, versionName, description string) (AgentVersion, error) {
	payload := map[string]any{
		"input_definitions": inputDefs,
		"engine_config":     engineConfig,
	}
	if versionName != "" {
		payload["version_name"] = versionName
	}
	if description != "" {
		payload["description"] = description
	}
	var respID struct {
		ID string `json:"id"`
	}
	if err := v.agentsAPI.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/agents/%s/versions/", agentID), payload, nil, &respID); err != nil {
		return AgentVersion{}, err
	}
	return v.RetrieveWithContext(ctx, agentID, respID.ID, nil)
}

func (v *AgentVersionsAPI) Update(agentID, versionID, versionName, description string) error {
	return v.UpdateWithContext(context.Background(), agentID, versionID, versionName, description)
}

func (v *AgentVersionsAPI) UpdateWithContext(ctx context.Context, agentID, versionID, versionName, description string) error {
	payload := map[string]any{}
	if versionName != "" {
		payload["version_name"] = versionName
	}
	if description != "" {
		payload["description"] = description
	}
	return v.agentsAPI.httpClient.putJSONWithContext(ctx, fmt.Sprintf("/v1/agents/%s/versions/%s/", agentID, versionID), payload, nil, nil)
}

func (v *AgentVersionsAPI) Delete(agentID, versionID string) error {
	return v.DeleteWithContext(context.Background(), agentID, versionID)
}

func (v *AgentVersionsAPI) DeleteWithContext(ctx context.Context, agentID, versionID string) error {
	return v.agentsAPI.httpClient.deleteWithContext(ctx, fmt.Sprintf("/v1/agents/%s/versions/%s/", agentID, versionID), nil)
}

// AgentJobsAPI handles job operations.
type AgentJobsAPI struct {
	agentsAPI *AgentsAPI
}

func (j *AgentJobsAPI) RetrieveStatus(jobID string) (AgentJobStatus, error) {
	return j.RetrieveStatusWithContext(context.Background(), jobID)
}

func (j *AgentJobsAPI) RetrieveStatusWithContext(ctx context.Context, jobID string) (AgentJobStatus, error) {
	var resp AgentJobStatus
	if err := j.agentsAPI.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/agents/jobs/%s/status/", jobID), nil, &resp); err != nil {
		return AgentJobStatus{}, err
	}
	return resp, nil
}

func (j *AgentJobsAPI) RetrieveResult(jobID string) (AgentJobResult, error) {
	return j.RetrieveResultWithContext(context.Background(), jobID)
}

func (j *AgentJobsAPI) RetrieveResultWithContext(ctx context.Context, jobID string) (AgentJobResult, error) {
	var resp AgentJobResult
	if err := j.agentsAPI.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/agents/jobs/%s/result/", jobID), nil, &resp); err != nil {
		return AgentJobResult{}, err
	}
	return resp, nil
}

func (j *AgentJobsAPI) RetrieveStatusMany(jobIDs []string) ([]AgentJobStatusBatch, error) {
	return j.RetrieveStatusManyWithContext(context.Background(), jobIDs)
}

func (j *AgentJobsAPI) RetrieveStatusManyWithContext(ctx context.Context, jobIDs []string) ([]AgentJobStatusBatch, error) {
	if len(jobIDs) == 0 {
		return nil, nil
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
		payload := map[string]any{"job_ids": chunk}
		var resp []AgentJobStatusBatch
		if err := j.agentsAPI.httpClient.postJSONWithContext(ctx, "/v1/agents/jobs/statuses/", payload, nil, &resp); err != nil {
			return nil, fmt.Errorf("retrieve job statuses: %w", err)
		}
		for _, st := range resp {
			if idx, ok := order[st.ID]; ok {
				results[idx] = st
				received[st.ID] = true
			}
		}
	}

	// Check for missing job IDs in response
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

func (j *AgentJobsAPI) RetrieveResultMany(jobIDs []string) ([]AgentJobResultBatch, error) {
	return j.RetrieveResultManyWithContext(context.Background(), jobIDs)
}

func (j *AgentJobsAPI) RetrieveResultManyWithContext(ctx context.Context, jobIDs []string) ([]AgentJobResultBatch, error) {
	if len(jobIDs) == 0 {
		return nil, nil
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
		payload := map[string]any{"job_ids": chunk}
		var resp []AgentJobResultBatch
		if err := j.agentsAPI.httpClient.postJSONWithContext(ctx, "/v1/agents/jobs/results/", payload, nil, &resp); err != nil {
			return nil, fmt.Errorf("retrieve job results: %w", err)
		}
		for _, st := range resp {
			if idx, ok := order[st.ID]; ok {
				results[idx] = st
				received[st.ID] = true
			}
		}
	}

	// Check for missing job IDs in response
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

func (j *AgentJobsAPI) DownloadReference(jobID, resourceID string, asAttachment bool) ([]byte, error) {
	return j.DownloadReferenceWithContext(context.Background(), jobID, resourceID, asAttachment)
}

func (j *AgentJobsAPI) DownloadReferenceWithContext(ctx context.Context, jobID, resourceID string, asAttachment bool) ([]byte, error) {
	params := map[string]string{}
	if asAttachment {
		params["download"] = "true"
	}
	return j.agentsAPI.httpClient.getBytesWithContext(ctx, fmt.Sprintf("/v1/agents/jobs/%s/references/%s/", jobID, resourceID), params)
}

func (j *AgentJobsAPI) DeleteData(jobID string) (JobDataDeleteResponse, error) {
	return j.DeleteDataWithContext(context.Background(), jobID)
}

func (j *AgentJobsAPI) DeleteDataWithContext(ctx context.Context, jobID string) (JobDataDeleteResponse, error) {
	var resp JobDataDeleteResponse
	if err := j.agentsAPI.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/agents/jobs/%s/delete-data/", jobID), nil, nil, &resp); err != nil {
		return JobDataDeleteResponse{}, err
	}
	return resp, nil
}

// helpers
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
