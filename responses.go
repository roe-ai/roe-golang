package roe

import (
	"encoding/json"
	"strings"
)

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
