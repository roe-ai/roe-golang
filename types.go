package roe

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AgentInputDefinition describes an agent input.
type AgentInputDefinition struct {
	Key                  string `json:"key"`
	DataType             string `json:"data_type"`
	Description          string `json:"description"`
	Example              string `json:"example,omitempty"`
	AcceptsMultipleFiles bool   `json:"accepts_multiple_files,omitempty"`
}

// UserInfo holds creator metadata.
type UserInfo struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// BaseAgent represents a base agent resource.
type BaseAgent struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Creator          *UserInfo  `json:"creator,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	DisableCache     bool       `json:"disable_cache"`
	CacheFailedJobs  bool       `json:"cache_failed_jobs"`
	OrganizationID   string     `json:"organization_id"`
	EngineClassID    string     `json:"engine_class_id"`
	CurrentVersionID *string    `json:"current_version_id"`
	JobCount         int        `json:"job_count"`
	MostRecentJob    *time.Time `json:"most_recent_job"`
	EngineName       string     `json:"engine_name"`

	agentsAPI *AgentsAPI `json:"-"`
}

func (a *BaseAgent) setAgentsAPI(api *AgentsAPI) {
	a.agentsAPI = api
}

// Run executes the agent using its current version.
func (a *BaseAgent) Run(inputs map[string]any) (*Job, error) {
	return a.RunWithContext(context.Background(), inputs)
}

// RunWithContext executes the agent using its current version with a caller-supplied context.
func (a *BaseAgent) RunWithContext(ctx context.Context, inputs map[string]any) (*Job, error) {
	if a.agentsAPI == nil {
		return nil, fmt.Errorf("agents API not set; use client.Agents.Run instead")
	}
	return a.agentsAPI.RunWithContext(ctx, a.ID, 0, inputs)
}

// ListVersions lists versions of this agent.
func (a *BaseAgent) ListVersions() ([]AgentVersion, error) {
	return a.ListVersionsWithContext(context.Background())
}

// ListVersionsWithContext lists versions of this agent with a caller-supplied context.
func (a *BaseAgent) ListVersionsWithContext(ctx context.Context) ([]AgentVersion, error) {
	if a.agentsAPI == nil {
		return nil, fmt.Errorf("agents API not set")
	}
	return a.agentsAPI.Versions.ListWithContext(ctx, a.ID)
}

// GetCurrentVersion retrieves current version.
func (a *BaseAgent) GetCurrentVersion() (*AgentVersion, error) {
	return a.GetCurrentVersionWithContext(context.Background())
}

// GetCurrentVersionWithContext retrieves current version with a caller-supplied context.
func (a *BaseAgent) GetCurrentVersionWithContext(ctx context.Context) (*AgentVersion, error) {
	if a.agentsAPI == nil {
		return nil, fmt.Errorf("agents API not set")
	}
	if a.CurrentVersionID == nil {
		return nil, nil
	}
	v, err := a.agentsAPI.Versions.RetrieveWithContext(ctx, a.ID, *a.CurrentVersionID, nil)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// AgentVersion describes a specific agent version.
type AgentVersion struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	VersionName    string                 `json:"version_name"`
	Creator        *UserInfo              `json:"creator,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	Description    *string                `json:"description"`
	EngineClassID  string                 `json:"engine_class_id"`
	EngineName     string                 `json:"engine_name"`
	InputDefs      []AgentInputDefinition `json:"input_definitions"`
	EngineConfig   map[string]any         `json:"engine_config"`
	OrganizationID string                 `json:"organization_id"`
	Readonly       bool                   `json:"readonly"`
	BaseAgent      BaseAgent              `json:"base_agent"`

	agentsAPI *AgentsAPI `json:"-"`
}

func (v *AgentVersion) setAgentsAPI(api *AgentsAPI) {
	v.agentsAPI = api
	v.BaseAgent.setAgentsAPI(api)
}

// Run executes this version directly.
func (v *AgentVersion) Run(inputs map[string]any) (*Job, error) {
	return v.RunWithContext(context.Background(), inputs)
}

// RunWithContext executes this version directly with a caller-supplied context.
func (v *AgentVersion) RunWithContext(ctx context.Context, inputs map[string]any) (*Job, error) {
	if v.agentsAPI == nil {
		return nil, fmt.Errorf("agents API not set; use client.Agents.Run instead")
	}
	return v.agentsAPI.RunVersionWithContext(ctx, v.BaseAgent.ID, v.ID, 0, inputs)
}

// JobStatus enum values.
type JobStatus int

const (
	JobPending JobStatus = iota
	JobStarted
	JobRetry
	JobSuccess
	JobFailure
	JobCancelled
	JobCached
)

func (s JobStatus) String() string {
	switch s {
	case JobPending:
		return "pending"
	case JobStarted:
		return "started"
	case JobRetry:
		return "retry"
	case JobSuccess:
		return "success"
	case JobFailure:
		return "failure"
	case JobCancelled:
		return "cancelled"
	case JobCached:
		return "cached"
	default:
		return "unknown"
	}
}

func (s JobStatus) IsTerminal() bool {
	return s == JobSuccess || s == JobFailure || s == JobCancelled || s == JobCached
}

// AgentDatum represents a job output.
type AgentDatum struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	DataType    string   `json:"data_type"`
	Value       string   `json:"value"`
	Cost        *float64 `json:"cost,omitempty"`
}

// PaginatedResponse wraps paginated results.
type PaginatedResponse[T any] struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []T     `json:"results"`
}

func (p PaginatedResponse[T]) HasNext() bool {
	return p.Next != nil
}

func (p PaginatedResponse[T]) HasPrevious() bool {
	return p.Previous != nil
}

// AgentJobStatus describes job status.
type AgentJobStatus struct {
	Status       JobStatus `json:"status"`
	Timestamp    float64   `json:"timestamp"`
	ErrorMessage *string   `json:"error_message"`
}

// Reference describes downloadable reference file.
type Reference struct {
	URL        string `json:"url"`
	ResourceID string `json:"resource_id"`
}

// AgentJobResult describes job result payload.
type AgentJobResult struct {
	AgentID        string       `json:"agent_id"`
	AgentVersionID string       `json:"agent_version_id"`
	Inputs         []any        `json:"inputs"`
	InputTokens    *int         `json:"input_tokens"`
	OutputTokens   *int         `json:"output_tokens"`
	Outputs        []AgentDatum `json:"outputs"`
}

// GetReferences parses outputs for reference URLs.
func (r AgentJobResult) GetReferences() []Reference {
	var refs []Reference
	for _, out := range r.Outputs {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(out.Value), &parsed); err != nil {
			continue
		}
		raw, ok := parsed["references"]
		if !ok {
			continue
		}
		arr, ok := raw.([]any)
		if !ok {
			continue
		}
		for _, item := range arr {
			s, ok := item.(string)
			if !ok {
				continue
			}
			if !strings.Contains(s, "/references/") {
				continue
			}
			parts := strings.Split(s, "/references/")
			if len(parts) > 1 {
				resource := strings.TrimSuffix(parts[len(parts)-1], "/")
				refs = append(refs, Reference{URL: s, ResourceID: resource})
			}
		}
	}
	return refs
}

// AgentJobStatusBatch for batch status.
type AgentJobStatusBatch struct {
	ID            string     `json:"id"`
	Status        *JobStatus `json:"status"`
	CreatedAt     any        `json:"created_at"`
	LastUpdatedAt any        `json:"last_updated_at"`
}

// AgentJobResultBatch for batch results.
type AgentJobResultBatch struct {
	ID             string       `json:"id"`
	Status         *JobStatus   `json:"status"`
	Result         any          `json:"result"`
	Corrected      []AgentDatum `json:"corrected_outputs,omitempty"`
	AgentID        *string      `json:"agent_id"`
	AgentVersionID *string      `json:"agent_version_id"`
	Cost           *float64     `json:"cost"`
	Inputs         []any        `json:"inputs"`
	InputTokens    *int         `json:"input_tokens"`
	OutputTokens   *int         `json:"output_tokens"`
}

// JobDataDeleteResponse for delete-data endpoint.
type JobDataDeleteResponse struct {
	Status           string   `json:"status"`
	DeletedCount     int      `json:"deleted_count"`
	FailedCount      int      `json:"failed_count"`
	OutputsSanitized bool     `json:"outputs_sanitized"`
	Errors           []string `json:"errors"`
}
