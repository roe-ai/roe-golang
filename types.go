package roe

import (
	"encoding/json"
	"strings"
)

// NOTE: Hand-written response models (BaseAgent, AgentVersion, Policy,
// PolicyVersion, AgentInputDefinition, UserInfo, PaginatedResponse[T]) have
// been removed; the AgentsAPI / PoliciesAPI wrappers now return generated
// equivalents from the github.com/roe-ai/roe-golang/generated package.
//
// What remains in this file is the polling-helper transport layer that
// Job/JobBatch use internally: AgentJobStatus, AgentJobStatusBatch,
// AgentJobResult, AgentJobResultBatch, AgentDatum, Reference, JobStatus
// (typed enum + IsTerminal()), and JobDataDeleteResponse. The bulk-status
// and bulk-result wire formats include per-item job IDs that the generated
// schemas omit, so these structs continue to back the parse path.

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
	Status         *JobStatus   `json:"status,omitempty"`
	ErrorMessage   *string      `json:"error_message,omitempty"`
}

// Succeeded returns true if the job status is SUCCESS or CACHED.
// Note: Status is only populated by WaitContext; direct RetrieveResult calls leave it nil.
func (r AgentJobResult) Succeeded() bool {
	if r.Status == nil {
		return false
	}
	return *r.Status == JobSuccess || *r.Status == JobCached
}

// Failed returns true if the job status is FAILURE or CANCELLED.
// Note: Status is only populated by WaitContext; direct RetrieveResult calls leave it nil.
func (r AgentJobResult) Failed() bool {
	if r.Status == nil {
		return false
	}
	return *r.Status == JobFailure || *r.Status == JobCancelled
}

// Cancelled returns true if the job status is CANCELLED.
// Note: Status is only populated by WaitContext; direct RetrieveResult calls leave it nil.
func (r AgentJobResult) Cancelled() bool {
	if r.Status == nil {
		return false
	}
	return *r.Status == JobCancelled
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
	Timestamp     *float64   `json:"timestamp"`
	ErrorMessage  *string    `json:"error_message"`
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

// Policy represents a policy resource for agentic workflows.
type Policy struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	OrganizationID   string  `json:"organization_id"`
	CurrentVersionID *string `json:"current_version_id"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// PolicyVersion represents a specific version of a policy.
type PolicyVersion struct {
	ID            string         `json:"id"`
	VersionName   string         `json:"version_name"`
	Content       map[string]any `json:"content"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	Policy        *Policy        `json:"policy,omitempty"`
	CreatedBy     map[string]any `json:"created_by,omitempty"`
	BaseVersionID *string        `json:"base_version_id,omitempty"`
}
