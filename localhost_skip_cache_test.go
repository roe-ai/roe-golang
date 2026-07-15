package roe_test

// Manual end-to-end check of the new skip-cache option against a local dev
// stack (./roe-cli dev). Not meant for CI — it only runs when
// ROE_LOCAL_API_KEY is set. Do not commit.
//
// Usage:
//
//	cd roe-golang
//	ROE_LOCAL_API_KEY=<local api key> go test -v -timeout 40m -run TestSkipCacheAgainstLocalhost .
//
// Optional env overrides:
//
//	ROE_LOCAL_BASE_URL      default http://localhost:8000/api
//	ROE_LOCAL_INPUT         merchant_context value, default below (same value
//	                        must be reused across runs for the cache to hit)
//	ROE_LOCAL_KEEP_JOB=1    let the skip-cache run finish instead of cancelling
//
// What it verifies, in order:
//  1. A run without the option completes (or is served from cache if this
//     input was already run — both are fine as the warm-cache baseline).
//  2. A second run with the same input WITHOUT skip-cache returns the SAME
//     job id, served as CACHED — proving the cache is live for this input.
//  3. A run WITH roe.RunOptions{SkipCache: true} gets a NEW job id and is not served
//     from cache. It is then cancelled to save compute (like the cancelled
//     test runs in the agent's case list).

import (
	"os"
	"testing"
	"time"

	roe "github.com/roe-ai/roe-golang"
)

const (
	localAgentID = "44b80cc0-3ff3-410d-a906-34ca92964de9" // TT <> Whatnot Eval (Merchant Risk)
	localOrgID   = "26ea8998-b382-4785-bcff-f5c1f62d91ba"

	// Merchant Risk runs take ~5-8 min on this agent; leave headroom.
	freshRunTimeout = 30 * time.Minute
	pollInterval    = 10 * time.Second
)

func TestSkipCacheAgainstLocalhost(t *testing.T) {
	apiKey := os.Getenv("ROE_LOCAL_API_KEY")
	if apiKey == "" {
		t.Skip("set ROE_LOCAL_API_KEY to run this manual localhost test")
	}
	baseURL := os.Getenv("ROE_LOCAL_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000/api"
	}
	merchantContext := os.Getenv("ROE_LOCAL_INPUT")
	if merchantContext == "" {
		merchantContext = "Zenmart - skip-cache SDK test"
	}

	client, err := roe.NewClientWithConfig(roe.Config{
		APIKey:         apiKey,
		OrganizationID: localOrgID,
		BaseURL:        baseURL,
		Timeout:        2 * time.Minute,
		MaxRetries:     1,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	inputs := map[string]any{"merchant_context": merchantContext}

	// ---- Phase 1: baseline run, no skip-cache. Warms the cache (or hits it).
	t.Logf("phase 1: submitting baseline run (input=%q)", merchantContext)
	start := time.Now()
	job1, err := client.Agents.Run(localAgentID, 0, inputs, nil)
	if err != nil {
		t.Fatalf("phase 1 run: %v", err)
	}
	t.Logf("phase 1: job %s submitted in %s", job1.ID(), time.Since(start).Round(time.Millisecond))

	status1, err := job1.RetrieveStatus()
	if err != nil {
		t.Fatalf("phase 1 status: %v", err)
	}
	if status1.Status == roe.JobCached {
		t.Logf("phase 1: input was already cached (job %s served as CACHED) — using it as the baseline", job1.ID())
	} else {
		t.Logf("phase 1: fresh run (status %s), waiting up to %s for completion...", status1.Status, freshRunTimeout)
		result, err := job1.Wait(pollInterval, freshRunTimeout)
		if err != nil {
			t.Fatalf("phase 1 wait: %v", err)
		}
		if result.Status == nil || (*result.Status != roe.JobSuccess && *result.Status != roe.JobCached) {
			t.Fatalf("phase 1: baseline run did not succeed (status %v) — cannot test caching against a failed job", result.Status)
		}
		t.Logf("phase 1: baseline completed in %s", time.Since(start).Round(time.Second))
	}

	// ---- Phase 2: same input, still no skip-cache. Must be served from cache.
	t.Log("phase 2: re-running same input WITHOUT skip-cache (expect cache hit)")
	start = time.Now()
	job2, err := client.Agents.Run(localAgentID, 0, inputs, nil)
	if err != nil {
		t.Fatalf("phase 2 run: %v", err)
	}
	elapsed2 := time.Since(start)
	if job2.ID() != job1.ID() {
		t.Errorf("phase 2: expected cache to return baseline job %s, got new job %s — cache lookup did NOT hit", job1.ID(), job2.ID())
	}
	status2, err := job2.RetrieveStatus()
	if err != nil {
		t.Fatalf("phase 2 status: %v", err)
	}
	if status2.Status != roe.JobCached {
		t.Errorf("phase 2: expected CACHED status, got %s", status2.Status)
	} else {
		t.Logf("phase 2: cache hit confirmed — same job id, status CACHED, returned in %s", elapsed2.Round(time.Millisecond))
	}

	// ---- Phase 3: same input WITH skip-cache. Must be a brand-new job.
	t.Log("phase 3: re-running same input WITH roe.RunOptions{SkipCache: true} (expect fresh job)")
	start = time.Now()
	job3, err := client.Agents.Run(localAgentID, 0, inputs, nil, roe.RunOptions{SkipCache: true})
	if err != nil {
		t.Fatalf("phase 3 run: %v", err)
	}
	elapsed3 := time.Since(start)
	if job3.ID() == job1.ID() {
		t.Fatalf("phase 3: skip-cache returned the cached job %s — X-Skip-Cache was NOT honored", job1.ID())
	}
	status3, err := job3.RetrieveStatus()
	if err != nil {
		t.Fatalf("phase 3 status: %v", err)
	}
	if status3.Status == roe.JobCached {
		t.Errorf("phase 3: new job %s reported CACHED immediately — unexpected", job3.ID())
	} else {
		t.Logf("phase 3: fresh job %s (status %s) submitted in %s — skip-cache honored",
			job3.ID(), status3.Status, elapsed3.Round(time.Millisecond))
	}

	if os.Getenv("ROE_LOCAL_KEEP_JOB") == "1" {
		t.Logf("phase 3: leaving job %s running (ROE_LOCAL_KEEP_JOB=1); it will refresh the cache on completion", job3.ID())
	} else {
		if err := client.Agents.Jobs.Cancel(job3.ID()); err != nil {
			t.Logf("phase 3: cancel of fresh job %s failed (non-fatal): %v", job3.ID(), err)
		} else {
			t.Logf("phase 3: fresh job %s cancelled to save compute", job3.ID())
		}
	}

	t.Log("PASS criteria: phase 2 same-id+CACHED, phase 3 new-id+not-CACHED")
}
