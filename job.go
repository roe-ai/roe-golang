package roe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Job represents a single agent job.
type Job struct {
	agentsAPI *AgentsAPI
	jobID     string
	timeout   time.Duration
}

func newJob(api *AgentsAPI, jobID string, timeoutSeconds int) *Job {
	to := 7200 * time.Second
	if timeoutSeconds > 0 {
		to = time.Duration(timeoutSeconds) * time.Second
	}
	return &Job{agentsAPI: api, jobID: jobID, timeout: to}
}

func (j *Job) ID() string {
	return j.jobID
}

func (j *Job) Timeout() time.Duration {
	return j.timeout
}

// Wait polls for completion and returns result.
func (j *Job) Wait(interval time.Duration, timeout time.Duration) (AgentJobResult, error) {
	return j.WaitContext(context.Background(), interval, timeout)
}

// WaitContext polls for completion with a caller-supplied context.
func (j *Job) WaitContext(ctx context.Context, interval time.Duration, timeout time.Duration) (AgentJobResult, error) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = j.timeout
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return AgentJobResult{}, fmt.Errorf("job %s wait cancelled: %w", j.jobID, ctx.Err())
		default:
		}

		status, err := j.RetrieveStatusWithContext(ctx)
		if err != nil {
			return AgentJobResult{}, err
		}
		if status.Status.IsTerminal() {
			result, err := j.RetrieveResultWithContext(ctx)
			if err != nil {
				return AgentJobResult{}, err
			}
			if status.Status == JobFailure || status.Status == JobCancelled {
				return result, fmt.Errorf("job %s ended with status %s", j.jobID, status.Status.String())
			}
			return result, nil
		}

		select {
		case <-ctx.Done():
			return AgentJobResult{}, fmt.Errorf("job %s wait cancelled: %w", j.jobID, ctx.Err())
		case <-ticker.C:
		}
	}
}

// RetrieveStatus fetches job status.
func (j *Job) RetrieveStatus() (AgentJobStatus, error) {
	if j.agentsAPI == nil {
		return AgentJobStatus{}, errors.New("agents API not set")
	}
	return j.agentsAPI.Jobs.RetrieveStatus(j.jobID)
}

// RetrieveStatusWithContext fetches job status with a provided context.
func (j *Job) RetrieveStatusWithContext(ctx context.Context) (AgentJobStatus, error) {
	if j.agentsAPI == nil {
		return AgentJobStatus{}, errors.New("agents API not set")
	}
	return j.agentsAPI.Jobs.RetrieveStatusWithContext(ctx, j.jobID)
}

// GetStatus is an alias for RetrieveStatus (kept for parity with examples).
func (j *Job) GetStatus() (AgentJobStatus, error) {
	return j.RetrieveStatus()
}

// RetrieveResult fetches job result.
func (j *Job) RetrieveResult() (AgentJobResult, error) {
	if j.agentsAPI == nil {
		return AgentJobResult{}, errors.New("agents API not set")
	}
	return j.agentsAPI.Jobs.RetrieveResult(j.jobID)
}

// RetrieveResultWithContext fetches job result with a provided context.
func (j *Job) RetrieveResultWithContext(ctx context.Context) (AgentJobResult, error) {
	if j.agentsAPI == nil {
		return AgentJobResult{}, errors.New("agents API not set")
	}
	return j.agentsAPI.Jobs.RetrieveResultWithContext(ctx, j.jobID)
}

// JobBatch tracks multiple jobs.
type JobBatch struct {
	agentsAPI *AgentsAPI
	jobIDs    []string
	timeout   time.Duration
	statuses  map[string]JobStatus
	completed map[string]AgentJobResult
}

func newJobBatch(api *AgentsAPI, jobIDs []string, timeoutSeconds int) *JobBatch {
	to := 7200 * time.Second
	if timeoutSeconds > 0 {
		to = time.Duration(timeoutSeconds) * time.Second
	}
	return &JobBatch{
		agentsAPI: api,
		jobIDs:    jobIDs,
		timeout:   to,
		statuses:  map[string]JobStatus{},
		completed: map[string]AgentJobResult{},
	}
}

// Jobs returns individual Job handles.
func (b *JobBatch) Jobs() []*Job {
	jobs := make([]*Job, 0, len(b.jobIDs))
	for _, id := range b.jobIDs {
		jobs = append(jobs, newJob(b.agentsAPI, id, int(b.timeout/time.Second)))
	}
	return jobs
}

// Wait waits for all jobs to finish and returns results in same order.
func (b *JobBatch) Wait(interval time.Duration, timeout time.Duration) ([]AgentJobResult, error) {
	return b.WaitContext(context.Background(), interval, timeout)
}

// WaitContext waits for all jobs with context cancellation and ordered results.
func (b *JobBatch) WaitContext(ctx context.Context, interval time.Duration, timeout time.Duration) ([]AgentJobResult, error) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = b.timeout
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	pending := append([]string{}, b.jobIDs...)
	failures := map[string]JobStatus{}

	for len(pending) > 0 {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("job batch wait cancelled: %w", ctx.Err())
		default:
		}

		statusBatch, err := b.agentsAPI.Jobs.RetrieveStatusManyWithContext(ctx, pending)
		if err != nil {
			return nil, err
		}

		var ready []string
		for _, st := range statusBatch {
			if st.Status != nil {
				b.statuses[st.ID] = *st.Status
				if st.Status.IsTerminal() {
					ready = append(ready, st.ID)
				}
			}
		}

		if len(ready) > 0 {
			resultsBatch, err := b.agentsAPI.Jobs.RetrieveResultManyWithContext(ctx, ready)
			if err != nil {
				return nil, err
			}

			received := map[string]AgentJobResult{}
			for _, res := range resultsBatch {
				converted, err := convertBatchResult(res)
				if err != nil {
					return nil, err
				}
				received[res.ID] = converted
				b.completed[res.ID] = converted
				if status, ok := b.statuses[res.ID]; ok && (status == JobFailure || status == JobCancelled) {
					failures[res.ID] = status
				}
			}

			for _, id := range ready {
				if _, ok := received[id]; !ok {
					return nil, fmt.Errorf("job %s result missing in batch response", id)
				}
			}

			pending = removeCompleted(pending, ready)
		}

		if len(pending) == 0 {
			break
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("job batch wait cancelled: %w", ctx.Err())
		case <-ticker.C:
		}
	}

	results := make([]AgentJobResult, 0, len(b.jobIDs))
	for _, id := range b.jobIDs {
		if res, ok := b.completed[id]; ok {
			results = append(results, res)
		}
	}

	if len(failures) > 0 {
		return results, fmt.Errorf("one or more jobs failed or were cancelled: %v", mapKeys(failures))
	}

	return results, nil
}

// RetrieveStatus returns latest known statuses keyed by job id.
func (b *JobBatch) RetrieveStatus() (map[string]JobStatus, error) {
	statusMap := map[string]JobStatus{}
	toQuery := []string{}
	for _, id := range b.jobIDs {
		if st, ok := b.statuses[id]; ok {
			statusMap[id] = st
		} else {
			toQuery = append(toQuery, id)
		}
	}
	if len(toQuery) > 0 {
		statusBatch, err := b.agentsAPI.Jobs.RetrieveStatusMany(toQuery)
		if err != nil {
			return nil, err
		}
		for _, st := range statusBatch {
			if st.Status != nil {
				statusMap[st.ID] = *st.Status
				b.statuses[st.ID] = *st.Status
			}
		}
	}
	return statusMap, nil
}

func removeCompleted(pending []string, completed []string) []string {
	if len(completed) == 0 {
		return pending
	}
	completedSet := map[string]struct{}{}
	for _, id := range completed {
		completedSet[id] = struct{}{}
	}
	next := make([]string, 0, len(pending))
	for _, id := range pending {
		if _, done := completedSet[id]; !done {
			next = append(next, id)
		}
	}
	return next
}

func mapKeys(m map[string]JobStatus) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func convertBatchResult(res AgentJobResultBatch) (AgentJobResult, error) {
	if res.AgentID == nil || res.AgentVersionID == nil {
		return AgentJobResult{}, fmt.Errorf("job %s not found or deleted", res.ID)
	}

	outputs := []AgentDatum{}

	converted := AgentJobResult{
		AgentID:        *res.AgentID,
		AgentVersionID: *res.AgentVersionID,
		Inputs:         res.Inputs,
		InputTokens:    res.InputTokens,
		OutputTokens:   res.OutputTokens,
	}

	switch data := res.Result.(type) {
	case []any:
		for i, raw := range data {
			datumBytes, err := json.Marshal(raw)
			if err != nil {
				return AgentJobResult{}, fmt.Errorf("job %s: marshal output[%d]: %w", res.ID, i, err)
			}
			var datum AgentDatum
			if err := json.Unmarshal(datumBytes, &datum); err != nil {
				return AgentJobResult{}, fmt.Errorf("job %s: unmarshal output[%d]: %w", res.ID, i, err)
			}
			outputs = append(outputs, datum)
		}
	case nil:
		// No result data - outputs remains empty
	default:
		// Try to unmarshal as array of AgentDatum
		var direct []AgentDatum
		bytes, err := json.Marshal(res.Result)
		if err != nil {
			return AgentJobResult{}, fmt.Errorf("job %s: marshal result: %w", res.ID, err)
		}
		if err := json.Unmarshal(bytes, &direct); err != nil {
			return AgentJobResult{}, fmt.Errorf("job %s: unmarshal result as []AgentDatum: %w", res.ID, err)
		}
		outputs = append(outputs, direct...)
	}

	if len(outputs) == 0 && len(res.Corrected) > 0 {
		outputs = res.Corrected
	}
	converted.Outputs = outputs

	return converted, nil
}
