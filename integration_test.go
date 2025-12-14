//go:build integration
// +build integration

package roe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

/*
Comprehensive integration tests for the Roe Go SDK.
Tests identical scenarios to Python and TypeScript SDKs for cross-SDK parity verification.

Run with: go test -tags=integration -v ./...

Note: These tests require network access and valid API credentials.
*/

// Test configuration
var testConfig = struct {
	APIKey         string
	OrganizationID string
	BaseURL        string
	SamplePDFs     []string
	SampleURL      string
}{
	APIKey:         "7jmZWLJO.QvCXpF1v7IhXebQw4TLoW5YFtG89jDJo",
	OrganizationID: "abfad6df-ee35-4ab0-81ab-89214e9ed900",
	BaseURL:        "https://api.test.roe-ai.com",
	SamplePDFs: []string{
		"https://www.uscis.gov/sites/default/files/document/forms/i-9.pdf",
		"https://arxiv.org/pdf/2512.09941",
		"https://arxiv.org/pdf/2512.09933",
		"https://arxiv.org/pdf/2512.09956",
	},
	SampleURL: "https://www.roe-ai.com/",
}

type TestResult map[string]interface{}

type TestResults struct {
	Results map[string]TestResult
	Errors  [][]string
}

func NewTestResults() *TestResults {
	return &TestResults{
		Results: make(map[string]TestResult),
		Errors:  make([][]string, 0),
	}
}

func (r *TestResults) Record(testName string, result TestResult) {
	r.Results[testName] = result
	fmt.Printf("  [PASS] %s\n", testName)
}

func (r *TestResults) RecordError(testName string, err error) {
	r.Errors = append(r.Errors, []string{testName, err.Error()})
	fmt.Printf("  [FAIL] %s: %v\n", testName, err)
}

func (r *TestResults) ToJSON() string {
	data, _ := json.MarshalIndent(map[string]interface{}{
		"results": r.Results,
		"errors":  r.Errors,
		"passed":  len(r.Results),
		"failed":  len(r.Errors),
	}, "", "  ")
	return string(data)
}

// downloadPDF downloads a PDF from URL to a temp file
func downloadPDF(url string, filename string) (string, error) {
	tempDir, err := os.MkdirTemp("", "roe-test-")
	if err != nil {
		return "", err
	}

	filepath := filepath.Join(tempDir, filename)
	fmt.Printf("    Downloading %s...\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Handle redirects
	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		redirectURL := resp.Header.Get("Location")
		if redirectURL != "" {
			os.RemoveAll(tempDir)
			return downloadPDF(redirectURL, filename)
		}
	}

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filepath, nil
}

// cleanupFile removes a temp file and its directory
func cleanupFile(filepath string) {
	os.Remove(filepath)
	os.RemoveAll(filepath[:len(filepath)-len("/"+filepath[len(filepath)-1:])])
}

// TestConfigEdgeCases tests configuration edge cases
func TestConfigEdgeCases(t *testing.T) {
	fmt.Println("\n=== Testing Config Edge Cases ===")
	results := NewTestResults()

	// Test 1: Zero values should work (where valid)
	// Note: Go SDK doesn't allow timeout=0 or maxRetries=0 as they use default fallback
	// This is a known difference from Python/TypeScript
	cfg, err := LoadConfigWithParams(ConfigParams{
		APIKey:         testConfig.APIKey,
		OrganizationID: testConfig.OrganizationID,
		BaseURL:        testConfig.BaseURL,
		TimeoutSeconds: 1.0, // Minimum positive value
		MaxRetries:     1,   // Go doesn't treat 0 specially
	})
	if err != nil {
		results.RecordError("config_low_values", err)
	} else {
		results.Record("config_low_values", TestResult{
			"timeout":     cfg.Timeout.Seconds(),
			"max_retries": cfg.MaxRetries,
		})
	}

	// Test 2: Default config creation
	cfg2, err := LoadConfigWithParams(ConfigParams{
		APIKey:         testConfig.APIKey,
		OrganizationID: testConfig.OrganizationID,
		BaseURL:        testConfig.BaseURL,
	})
	if err != nil {
		results.RecordError("config_defaults", err)
	} else {
		if cfg2.Timeout != 60*time.Second {
			results.RecordError("config_defaults", fmt.Errorf("expected default timeout=60s, got %v", cfg2.Timeout))
		} else if cfg2.MaxRetries != 3 {
			results.RecordError("config_defaults", fmt.Errorf("expected default max_retries=3, got %d", cfg2.MaxRetries))
		} else {
			results.Record("config_defaults", TestResult{
				"timeout":     cfg2.Timeout.Seconds(),
				"max_retries": cfg2.MaxRetries,
			})
		}
	}

	fmt.Printf("\nConfig Edge Cases: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// TestFileUploadFromPath tests FileUpload from local path
func TestFileUploadFromPath(t *testing.T) {
	fmt.Println("\n=== Testing FileUpload from Path ===")
	results := NewTestResults()

	pdfPath, err := downloadPDF(testConfig.SamplePDFs[1], "test_upload.pdf")
	if err != nil {
		results.RecordError("download_pdf", err)
		return
	}
	defer cleanupFile(pdfPath)

	// Test FileUpload with file path
	upload := FileUpload{Path: pdfPath}
	filename := upload.filename()
	mimeType := upload.mimeType()

	if filename != "test_upload.pdf" {
		results.RecordError("file_upload_from_path", fmt.Errorf("expected filename 'test_upload.pdf', got '%s'", filename))
	} else if mimeType != "application/pdf" {
		results.RecordError("file_upload_from_path", fmt.Errorf("expected mimeType 'application/pdf', got '%s'", mimeType))
	} else {
		results.Record("file_upload_from_path", TestResult{
			"filename":  filename,
			"mime_type": mimeType,
		})
	}

	fmt.Printf("\nFileUpload: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// TestDocInsightsAgent tests Document Insights (PDF Extraction) agent
func TestDocInsightsAgent(t *testing.T) {
	fmt.Println("\n=== Testing Doc Insights Agent ===")
	results := NewTestResults()

	client, err := NewClient(
		testConfig.APIKey,
		testConfig.OrganizationID,
		testConfig.BaseURL,
		60.0,
		3,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create a Doc Insights agent
	inputDefs := []map[string]any{
		{"key": "pdf_files", "data_type": "application/pdf", "description": "PDF document to analyze"},
	}
	engineConfig := map[string]any{
		"model":        "gpt-4.1-2025-04-14",
		"pdf_files":    "${pdf_files}",
		"instructions": "Extract the document title and list the main sections.",
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title":    map[string]any{"type": "string", "description": "Document title"},
				"sections": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Main sections"},
			},
		},
	}

	agent, err := client.Agents.Create("Go SDK Test - Doc Insights", "PDFExtractionEngine", inputDefs, engineConfig, "", "")
	if err != nil {
		results.RecordError("create_doc_insights_agent", err)
		fmt.Printf("\nDoc Insights Agent: %d passed, %d failed\n", len(results.Results), len(results.Errors))
		return
	}
	results.Record("create_doc_insights_agent", TestResult{"agent_id": agent.ID, "name": agent.Name})

	// Download and run with a PDF
	pdfPath, err := downloadPDF(testConfig.SamplePDFs[0], "i9_form.pdf")
	if err != nil {
		results.RecordError("download_pdf", err)
		client.Agents.Delete(agent.ID)
		return
	}

	// Run async job
	job, err := client.Agents.Run(agent.ID, 300, map[string]any{
		"pdf_files": FileUpload{Path: pdfPath},
	})
	if err != nil {
		results.RecordError("submit_doc_insights_job", err)
		cleanupFile(pdfPath)
		client.Agents.Delete(agent.ID)
		return
	}
	results.Record("submit_doc_insights_job", TestResult{"job_id": job.ID()})

	// Wait for result
	result, err := job.Wait(5*time.Second, 300*time.Second)
	cleanupFile(pdfPath)

	if err != nil {
		results.RecordError("doc_insights_result", err)
	} else {
		results.Record("doc_insights_result", TestResult{
			"agent_id":      result.AgentID,
			"outputs_count": len(result.Outputs),
			"has_output":    len(result.Outputs) > 0,
		})

		// Parse and verify output
		if len(result.Outputs) > 0 {
			outputValue := result.Outputs[0].Value
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(outputValue), &parsed); err == nil {
				_, hasTitle := parsed["title"]
				_, hasSections := parsed["sections"]
				results.Record("doc_insights_parsed", TestResult{
					"has_title":    hasTitle,
					"has_sections": hasSections,
				})
			} else {
				results.Record("doc_insights_parsed", TestResult{"raw_output": outputValue[:min(200, len(outputValue))]})
			}
		}
	}

	// Cleanup: delete the agent
	err = client.Agents.Delete(agent.ID)
	if err != nil {
		results.RecordError("delete_doc_insights_agent", err)
	} else {
		results.Record("delete_doc_insights_agent", TestResult{"deleted": true})
	}

	fmt.Printf("\nDoc Insights Agent: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// TestWebInsightsAgent tests Web Insights (URL Website Extraction) agent
func TestWebInsightsAgent(t *testing.T) {
	fmt.Println("\n=== Testing Web Insights Agent ===")
	results := NewTestResults()

	client, err := NewClient(
		testConfig.APIKey,
		testConfig.OrganizationID,
		testConfig.BaseURL,
		60.0,
		3,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create a Web Insights agent
	inputDefs := []map[string]any{
		{"key": "url", "data_type": "text/plain", "description": "Website URL to analyze"},
	}
	engineConfig := map[string]any{
		"url":         "${url}",
		"model":       "gpt-4.1-2025-04-14",
		"instruction": "Extract the company name and a brief description from this website.",
		"vision_mode": false,
		"crawl_config": map[string]any{
			"save_html":     true,
			"save_markdown": true,
		},
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"company_name": map[string]any{"type": "string", "description": "Company name"},
				"description":  map[string]any{"type": "string", "description": "Brief company description"},
			},
		},
	}

	agent, err := client.Agents.Create("Go SDK Test - Web Insights", "URLWebsiteExtractionEngine", inputDefs, engineConfig, "", "")
	if err != nil {
		results.RecordError("create_web_insights_agent", err)
		return
	}
	results.Record("create_web_insights_agent", TestResult{"agent_id": agent.ID, "name": agent.Name})

	// Run with URL
	job, err := client.Agents.Run(agent.ID, 300, map[string]any{
		"url": testConfig.SampleURL,
	})
	if err != nil {
		results.RecordError("submit_web_insights_job", err)
		client.Agents.Delete(agent.ID)
		return
	}
	results.Record("submit_web_insights_job", TestResult{"job_id": job.ID()})

	// Wait for result
	result, err := job.Wait(5*time.Second, 300*time.Second)
	if err != nil {
		results.RecordError("web_insights_result", err)
	} else {
		results.Record("web_insights_result", TestResult{
			"agent_id":      result.AgentID,
			"outputs_count": len(result.Outputs),
			"has_output":    len(result.Outputs) > 0,
		})

		// Parse and verify output
		if len(result.Outputs) > 0 {
			outputValue := result.Outputs[0].Value
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(outputValue), &parsed); err == nil {
				_, hasCompanyName := parsed["company_name"]
				_, hasDescription := parsed["description"]
				companyName := "N/A"
				if cn, ok := parsed["company_name"].(string); ok {
					companyName = cn
				}
				results.Record("web_insights_parsed", TestResult{
					"has_company_name": hasCompanyName,
					"has_description":  hasDescription,
					"company_name":     companyName,
				})
			} else {
				results.Record("web_insights_parsed", TestResult{"raw_output": outputValue[:min(200, len(outputValue))]})
			}
		}
	}

	// Cleanup
	err = client.Agents.Delete(agent.ID)
	if err != nil {
		results.RecordError("delete_web_insights_agent", err)
	} else {
		results.Record("delete_web_insights_agent", TestResult{"deleted": true})
	}

	fmt.Printf("\nWeb Insights Agent: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// TestBatchOperations tests batch job operations with multiple inputs
func TestBatchOperations(t *testing.T) {
	fmt.Println("\n=== Testing Batch Operations ===")
	results := NewTestResults()

	client, err := NewClient(
		testConfig.APIKey,
		testConfig.OrganizationID,
		testConfig.BaseURL,
		60.0,
		3,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create a simple text extraction agent for batch testing
	inputDefs := []map[string]any{
		{"key": "text", "data_type": "text/plain", "description": "Text to analyze"},
	}
	engineConfig := map[string]any{
		"model":       "gpt-4.1-2025-04-14",
		"text":        "${text}",
		"instruction": "Count the number of words in this text.",
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"word_count": map[string]any{"type": "integer", "description": "Number of words"},
			},
		},
	}

	agent, err := client.Agents.Create("Go SDK Test - Batch", "MultimodalExtractionEngine", inputDefs, engineConfig, "", "")
	if err != nil {
		results.RecordError("create_batch_agent", err)
		return
	}
	results.Record("create_batch_agent", TestResult{"agent_id": agent.ID})

	// Run batch with multiple inputs
	batchInputs := []map[string]any{
		{"text": "Hello world"},
		{"text": "This is a longer sentence with more words"},
		{"text": "One"},
	}

	batch, err := client.Agents.RunMany(agent.ID, batchInputs, 300)
	if err != nil {
		results.RecordError("submit_batch_jobs", err)
		client.Agents.Delete(agent.ID)
		return
	}
	results.Record("submit_batch_jobs", TestResult{"job_count": len(batch.Jobs())})

	// Wait for all results
	batchResults, err := batch.Wait(5*time.Second, 300*time.Second)
	if err != nil {
		results.RecordError("batch_results", err)
	} else {
		allHaveOutputs := true
		for _, r := range batchResults {
			if len(r.Outputs) == 0 {
				allHaveOutputs = false
				break
			}
		}
		results.Record("batch_results", TestResult{
			"results_count":    len(batchResults),
			"all_have_outputs": allHaveOutputs,
		})
	}

	// Cleanup
	err = client.Agents.Delete(agent.ID)
	if err != nil {
		results.RecordError("delete_batch_agent", err)
	} else {
		results.Record("delete_batch_agent", TestResult{"deleted": true})
	}

	fmt.Printf("\nBatch Operations: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// TestSyncExecution tests synchronous job execution
func TestSyncExecution(t *testing.T) {
	fmt.Println("\n=== Testing Sync Execution ===")
	results := NewTestResults()

	client, err := NewClient(
		testConfig.APIKey,
		testConfig.OrganizationID,
		testConfig.BaseURL,
		60.0,
		3,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create agent
	inputDefs := []map[string]any{
		{"key": "text", "data_type": "text/plain", "description": "Text input"},
	}
	engineConfig := map[string]any{
		"model":       "gpt-4.1-2025-04-14",
		"text":        "${text}",
		"instruction": "Echo back the input text.",
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"echo": map[string]any{"type": "string"},
			},
		},
	}

	agent, err := client.Agents.Create("Go SDK Test - Sync", "MultimodalExtractionEngine", inputDefs, engineConfig, "", "")
	if err != nil {
		results.RecordError("create_sync_agent", err)
		return
	}
	results.Record("create_sync_agent", TestResult{"agent_id": agent.ID})

	// Run synchronously
	outputs, err := client.Agents.RunSync(agent.ID, map[string]any{"text": "Test input for sync"})
	if err != nil {
		results.RecordError("sync_execution", err)
	} else {
		results.Record("sync_execution", TestResult{
			"outputs_count": len(outputs),
			"has_output":    len(outputs) > 0,
		})
	}

	// Cleanup
	err = client.Agents.Delete(agent.ID)
	if err != nil {
		results.RecordError("delete_sync_agent", err)
	} else {
		results.Record("delete_sync_agent", TestResult{"deleted": true})
	}

	fmt.Printf("\nSync Execution: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// TestVersionManagement tests agent version management
func TestVersionManagement(t *testing.T) {
	fmt.Println("\n=== Testing Version Management ===")
	results := NewTestResults()

	client, err := NewClient(
		testConfig.APIKey,
		testConfig.OrganizationID,
		testConfig.BaseURL,
		60.0,
		3,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create agent
	inputDefs := []map[string]any{
		{"key": "text", "data_type": "text/plain", "description": "Text"},
	}
	engineConfig := map[string]any{
		"model":       "gpt-4.1-2025-04-14",
		"text":        "${text}",
		"instruction": "Version 1 instruction",
		"output_schema": map[string]any{
			"type":       "object",
			"properties": map[string]any{"result": map[string]any{"type": "string"}},
		},
	}

	agent, err := client.Agents.Create("Go SDK Test - Versions", "MultimodalExtractionEngine", inputDefs, engineConfig, "v1", "")
	if err != nil {
		results.RecordError("create_versioned_agent", err)
		return
	}
	results.Record("create_versioned_agent", TestResult{"agent_id": agent.ID})

	// List versions
	versions, err := client.Agents.Versions.List(agent.ID)
	if err != nil {
		results.RecordError("list_versions", err)
	} else {
		results.Record("list_versions", TestResult{"version_count": len(versions)})
	}

	// Get current version
	current, err := client.Agents.Versions.RetrieveCurrent(agent.ID)
	if err != nil {
		results.RecordError("current_version", err)
	} else {
		results.Record("current_version", TestResult{
			"version_id":   current.ID,
			"version_name": current.VersionName,
		})
	}

	// Create new version
	v2Config := map[string]any{
		"model":       "gpt-4.1-2025-04-14",
		"text":        "${text}",
		"instruction": "Version 2 instruction - updated",
		"output_schema": map[string]any{
			"type":       "object",
			"properties": map[string]any{"result": map[string]any{"type": "string"}},
		},
	}
	v2, err := client.Agents.Versions.Create(agent.ID, []map[string]any{
		{"key": "text", "data_type": "text/plain", "description": "Text"},
	}, v2Config, "v2", "")
	if err != nil {
		results.RecordError("create_version_v2", err)
	} else {
		results.Record("create_version_v2", TestResult{
			"version_id":   v2.ID,
			"version_name": v2.VersionName,
		})

		// Run specific version
		job, err := client.Agents.RunVersion(agent.ID, v2.ID, 120, map[string]any{"text": "Test"})
		if err != nil {
			results.RecordError("run_specific_version", err)
		} else {
			result, err := job.Wait(5*time.Second, 120*time.Second)
			if err != nil {
				results.RecordError("run_specific_version", err)
			} else {
				results.Record("run_specific_version", TestResult{
					"job_id":     job.ID(),
					"has_output": len(result.Outputs) > 0,
				})
			}
		}
	}

	// Cleanup
	err = client.Agents.Delete(agent.ID)
	if err != nil {
		results.RecordError("delete_versioned_agent", err)
	} else {
		results.Record("delete_versioned_agent", TestResult{"deleted": true})
	}

	fmt.Printf("\nVersion Management: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// TestJobManagement tests job status and result retrieval
func TestJobManagement(t *testing.T) {
	fmt.Println("\n=== Testing Job Management ===")
	results := NewTestResults()

	client, err := NewClient(
		testConfig.APIKey,
		testConfig.OrganizationID,
		testConfig.BaseURL,
		60.0,
		3,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create agent
	inputDefs := []map[string]any{
		{"key": "text", "data_type": "text/plain", "description": "Text"},
	}
	engineConfig := map[string]any{
		"model":       "gpt-4.1-2025-04-14",
		"text":        "${text}",
		"instruction": "Analyze this text.",
		"output_schema": map[string]any{
			"type":       "object",
			"properties": map[string]any{"analysis": map[string]any{"type": "string"}},
		},
	}

	agent, err := client.Agents.Create("Go SDK Test - Jobs", "MultimodalExtractionEngine", inputDefs, engineConfig, "", "")
	if err != nil {
		results.RecordError("create_job_agent", err)
		return
	}

	// Submit job
	job, err := client.Agents.Run(agent.ID, 120, map[string]any{"text": "Test for job management"})
	if err != nil {
		results.RecordError("submit_job", err)
		client.Agents.Delete(agent.ID)
		return
	}
	results.Record("submit_job", TestResult{"job_id": job.ID()})

	// Check status
	status, err := client.Agents.Jobs.RetrieveStatus(job.ID())
	if err != nil {
		results.RecordError("job_status", err)
	} else {
		results.Record("job_status", TestResult{"status": status.Status})
	}

	// Wait and get result
	result, err := job.Wait(5*time.Second, 120*time.Second)
	if err != nil {
		results.RecordError("job_result", err)
	} else {
		results.Record("job_result", TestResult{
			"agent_id":    result.AgentID,
			"has_outputs": len(result.Outputs) > 0,
		})

		// Retrieve result separately
		retrievedResult, err := client.Agents.Jobs.RetrieveResult(job.ID())
		if err != nil {
			results.RecordError("retrieved_result", err)
		} else {
			results.Record("retrieved_result", TestResult{
				"matches": retrievedResult.AgentID == result.AgentID,
			})
		}
	}

	// Cleanup
	client.Agents.Delete(agent.ID)

	fmt.Printf("\nJob Management: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// TestMultiplePDFUploads tests uploading multiple different PDFs
func TestMultiplePDFUploads(t *testing.T) {
	fmt.Println("\n=== Testing Multiple PDF Uploads ===")
	results := NewTestResults()

	client, err := NewClient(
		testConfig.APIKey,
		testConfig.OrganizationID,
		testConfig.BaseURL,
		60.0,
		3,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Create agent
	inputDefs := []map[string]any{
		{"key": "pdf_files", "data_type": "application/pdf", "description": "PDF"},
	}
	engineConfig := map[string]any{
		"model":        "gpt-4.1-2025-04-14",
		"pdf_files":    "${pdf_files}",
		"instructions": "What is the title or main topic of this document?",
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{"type": "string"},
			},
		},
	}

	agent, err := client.Agents.Create("Go SDK Test - Multi PDF", "PDFExtractionEngine", inputDefs, engineConfig, "", "")
	if err != nil {
		results.RecordError("create_multi_pdf_agent", err)
		return
	}
	results.Record("create_multi_pdf_agent", TestResult{"agent_id": agent.ID})

	// Test with multiple PDFs (arxiv papers)
	pdfURLs := testConfig.SamplePDFs[1:4]
	for i, pdfURL := range pdfURLs {
		pdfPath, err := downloadPDF(pdfURL, fmt.Sprintf("arxiv_%d.pdf", i))
		if err != nil {
			results.RecordError(fmt.Sprintf("download_pdf_%d", i), err)
			continue
		}

		job, err := client.Agents.Run(agent.ID, 300, map[string]any{
			"pdf_files": FileUpload{Path: pdfPath},
		})
		if err != nil {
			results.RecordError(fmt.Sprintf("pdf_upload_%d", i), err)
			cleanupFile(pdfPath)
			continue
		}

		result, err := job.Wait(5*time.Second, 300*time.Second)
		cleanupFile(pdfPath)

		if err != nil {
			results.RecordError(fmt.Sprintf("pdf_upload_%d", i), err)
		} else {
			results.Record(fmt.Sprintf("pdf_upload_%d", i), TestResult{
				"url":        pdfURL,
				"job_id":     job.ID(),
				"has_output": len(result.Outputs) > 0,
			})
		}
	}

	// Cleanup
	err = client.Agents.Delete(agent.ID)
	if err != nil {
		results.RecordError("delete_multi_pdf_agent", err)
	} else {
		results.Record("delete_multi_pdf_agent", TestResult{"deleted": true})
	}

	fmt.Printf("\nMultiple PDF Uploads: %d passed, %d failed\n", len(results.Results), len(results.Errors))
}

// min returns the minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RunAllIntegrationTests runs all integration tests and saves results
func TestAllIntegration(t *testing.T) {
	fmt.Println("============================================================")
	fmt.Println("ROE GO SDK INTEGRATION TESTS")
	fmt.Println("============================================================")
	fmt.Printf("Base URL: %s\n", testConfig.BaseURL)
	fmt.Printf("Organization ID: %s\n", testConfig.OrganizationID)

	// Run individual test functions
	t.Run("ConfigEdgeCases", TestConfigEdgeCases)
	t.Run("FileUploadFromPath", TestFileUploadFromPath)
	t.Run("DocInsightsAgent", TestDocInsightsAgent)
	t.Run("WebInsightsAgent", TestWebInsightsAgent)
	t.Run("BatchOperations", TestBatchOperations)
	t.Run("SyncExecution", TestSyncExecution)
	t.Run("VersionManagement", TestVersionManagement)
	t.Run("JobManagement", TestJobManagement)
	t.Run("MultiplePDFUploads", TestMultiplePDFUploads)

	fmt.Println("\n============================================================")
	fmt.Println("TEST COMPLETE")
	fmt.Println("============================================================")
}
