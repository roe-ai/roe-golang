package roe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

// TablesAPI manages CSV uploads to Roe tables.
type TablesAPI struct {
	cfg        Config
	httpClient *httpClient
}

func newTablesAPI(cfg Config, httpClient *httpClient) *TablesAPI {
	return &TablesAPI{cfg: cfg, httpClient: httpClient}
}

// TableUploadResult is the parsed response from a successful upload.
//
// Summary is typed as `any` (matching the generated TableUploadResponse) so
// a future backend change that returns a non-object summary (null, string,
// array) does not break Unmarshal.
type TableUploadResult struct {
	TableName      string `json:"table_name"`
	OrganizationID string `json:"organization_id"`
	Summary        any    `json:"summary,omitempty"`
}

// TableUploadOptions controls optional upload parameters.
type TableUploadOptions struct {
	// WithHeaders indicates whether the first CSV row holds column names. Defaults to true.
	WithHeaders *bool
	// OrganizationID overrides the client default; must match the API key's org.
	OrganizationID string
}

// Upload uploads a CSV file (path or Reader) and creates a Roe table.
//
// FileUpload.URL is not supported on this surface — pass Path or Reader.
// Adding URL fetch here would couple the SDK to whatever fetcher the
// caller wants (auth, SSRF protections, retry policy); a caller who
// already has a URL should download it themselves and hand us a Reader.
func (t *TablesAPI) Upload(tableName string, file FileUpload, opts *TableUploadOptions) (TableUploadResult, error) {
	return t.UploadWithContext(context.Background(), tableName, file, opts)
}

// UploadWithContext uploads a CSV file with a caller-supplied context.
func (t *TablesAPI) UploadWithContext(ctx context.Context, tableName string, file FileUpload, opts *TableUploadOptions) (TableUploadResult, error) {
	if tableName == "" {
		return TableUploadResult{}, fmt.Errorf("tableName cannot be empty")
	}

	withHeaders := true
	orgID := t.cfg.OrganizationID
	if opts != nil {
		if opts.WithHeaders != nil {
			withHeaders = *opts.WithHeaders
		}
		if opts.OrganizationID != "" {
			orgID = opts.OrganizationID
		}
	}

	fileReader, filename, mimeType, err := t.httpClient.prepareMultipartFile(file)
	if err != nil {
		return TableUploadResult{}, err
	}
	defer fileReader.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("table_name", tableName); err != nil {
		return TableUploadResult{}, fmt.Errorf("write table_name field: %w", err)
	}
	if err := writer.WriteField("with_headers", fmt.Sprintf("%t", withHeaders)); err != nil {
		return TableUploadResult{}, fmt.Errorf("write with_headers field: %w", err)
	}
	if orgID != "" {
		if err := writer.WriteField("organization_id", orgID); err != nil {
			return TableUploadResult{}, fmt.Errorf("write organization_id field: %w", err)
		}
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name":     "file",
		"filename": filename,
	}))
	h.Set("Content-Type", mimeType)
	part, err := writer.CreatePart(h)
	if err != nil {
		writer.Close()
		return TableUploadResult{}, fmt.Errorf("create file part: %w", err)
	}
	if _, err := io.Copy(part, fileReader); err != nil {
		writer.Close()
		return TableUploadResult{}, fmt.Errorf("copy file body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return TableUploadResult{}, fmt.Errorf("close multipart writer: %w", err)
	}

	headers := http.Header{}
	headers.Set("Content-Type", writer.FormDataContentType())
	data, err := t.httpClient.doRequest(ctx, http.MethodPost, "/v1/tables/upload/", headers, body, nil)
	if err != nil {
		return TableUploadResult{}, err
	}
	var out TableUploadResult
	if err := json.Unmarshal(data, &out); err != nil {
		return TableUploadResult{}, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}
