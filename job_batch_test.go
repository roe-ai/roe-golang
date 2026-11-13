package roe

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestJobBatchWaitPreservesOrder(t *testing.T) {
	jobIDs := []string{"job-1", "job-2", "job-3"}

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/statuses/"):
			var payload struct {
				JobIDs []string `json:"job_ids"`
			}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			success := JobSuccess
			statuses := make([]AgentJobStatusBatch, 0, len(payload.JobIDs))
			for _, id := range payload.JobIDs {
				statuses = append(statuses, AgentJobStatusBatch{
					ID:     id,
					Status: &success,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(statuses)
		case strings.HasSuffix(r.URL.Path, "/results/"):
			agentID := "agent"
			versionID := "v1"
			results := []AgentJobResultBatch{
				{
					ID:             "job-2",
					Status:         nil,
					AgentID:        &agentID,
					AgentVersionID: &versionID,
					Result: []any{
						map[string]any{"key": "out", "value": "second", "description": "", "data_type": "text/plain"},
					},
				},
				{
					ID:             "job-1",
					Status:         nil,
					AgentID:        &agentID,
					AgentVersionID: &versionID,
					Result: []any{
						map[string]any{"key": "out", "value": "first", "description": "", "data_type": "text/plain"},
					},
				},
				{
					ID:             "job-3",
					Status:         nil,
					AgentID:        &agentID,
					AgentVersionID: &versionID,
					Result: []any{
						map[string]any{"key": "out", "value": "third", "description": "", "data_type": "text/plain"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(results)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{
		APIKey:               "k",
		OrganizationID:       "org",
		BaseURL:              server.URL,
		Timeout:              time.Second,
		MaxRetries:           0,
		RetryInitialInterval: 5 * time.Millisecond,
		RetryMaxInterval:     5 * time.Millisecond,
		RetryMultiplier:      1,
		RetryJitter:          0,
	}

	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()

	agents := newAgentsAPI(cfg, client)
	batch := newJobBatch(agents, jobIDs, 1)
	results, err := batch.Wait(5*time.Millisecond, time.Second)
	if err != nil {
		t.Fatalf("wait failed: %v", err)
	}
	if len(results) != len(jobIDs) {
		t.Fatalf("expected %d results, got %d", len(jobIDs), len(results))
	}

	values := []string{results[0].Outputs[0].Value, results[1].Outputs[0].Value, results[2].Outputs[0].Value}
	expected := []string{"first", "second", "third"}
	for i, v := range expected {
		if values[i] != v {
			t.Fatalf("result %d expected %s, got %s", i, v, values[i])
		}
	}
}

func TestConvertBatchResultAllowsErrorCodeOnFailedJob(t *testing.T) {
	status, agentID, versionID := JobFailure, "agent-id", "version-id"
	res, err := convertBatchResult(AgentJobResultBatch{
		ID:             "job-1",
		Status:         &status,
		Result:         "TIMEOUT",
		AgentID:        &agentID,
		AgentVersionID: &versionID,
	})
	if err != nil {
		t.Fatalf("convertBatchResult: %v", err)
	}
	if res.Status == nil || *res.Status != JobFailure || len(res.Outputs) != 0 {
		t.Fatalf("unexpected result %+v", res)
	}
}

// waitBatch runs JobBatch.Wait against a server that always answers the
// status and result endpoints with the given values.
func waitBatch(t *testing.T, jobIDs []string, statuses []AgentJobStatusBatch, results []AgentJobResultBatch, timeout time.Duration) ([]AgentJobResult, error) {
	t.Helper()
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/statuses/"):
			_ = json.NewEncoder(w).Encode(statuses)
		case strings.HasSuffix(r.URL.Path, "/results/"):
			_ = json.NewEncoder(w).Encode(results)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{APIKey: "k", OrganizationID: "org", BaseURL: server.URL, Timeout: time.Second, RetryMultiplier: 1}
	client := newHTTPClient(cfg, newAuth(cfg))
	defer client.close()
	return newJobBatch(newAgentsAPI(cfg, client), jobIDs, 1).Wait(5*time.Millisecond, timeout)
}

func TestJobBatchWaitUsesPolledStatusForFailedJobResult(t *testing.T) {
	success, failure := JobSuccess, JobFailure
	agentID, versionID := "agent", "v1"
	results, err := waitBatch(t, []string{"job-1", "job-2"},
		[]AgentJobStatusBatch{{ID: "job-1", Status: &success}, {ID: "job-2", Status: &failure}},
		[]AgentJobResultBatch{
			{ID: "job-1", AgentID: &agentID, AgentVersionID: &versionID, Result: []any{map[string]any{"key": "out", "value": "first"}}},
			// The results endpoint may leave status null for a failed job.
			{ID: "job-2", AgentID: &agentID, AgentVersionID: &versionID, Result: "TIMEOUT"},
		}, time.Second)
	if err != nil {
		t.Fatalf("wait failed: %v", err)
	}
	if len(results) != 2 || results[1].Status == nil || *results[1].Status != JobFailure || len(results[1].Outputs) != 0 {
		t.Fatalf("unexpected results %+v", results)
	}
}

func TestRunManyReturnsSubmittedJobsWhenALaterChunkFails(t *testing.T) {
	requests := 0
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests > 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ids := make([]string, maxBatchSize)
		for i := range ids {
			ids[i] = "job"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ids)
	}))
	defer server.Close()

	client := newAgentsTestClient(t, server.URL)
	defer client.Close()

	batch, err := client.Agents.RunMany("agent-id", make([]map[string]any, maxBatchSize+1), 0, nil)
	if err == nil || batch == nil || len(batch.Jobs()) != maxBatchSize {
		t.Fatalf("RunMany = %v, %v; want the first chunk's jobs and an error", batch, err)
	}
}
