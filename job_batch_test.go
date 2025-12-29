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
