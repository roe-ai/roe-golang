package roe

import (
	"context"
	"fmt"
)

// KnowledgeBaseAPI manages Knowledge Base lenses and drafts.
//
// This is a hand-maintained ("manual") wrapper: the knowledge_base operations
// are declared kind "manual" in the SDK contract, so codegen does not generate
// this file. It is wired explicitly in client.go like Agents/Policies/Users.
type KnowledgeBaseAPI struct {
	cfg        Config
	httpClient *httpClient
}

func newKnowledgeBaseAPI(cfg Config, httpClient *httpClient) *KnowledgeBaseAPI {
	return &KnowledgeBaseAPI{cfg: cfg, httpClient: httpClient}
}

// KnowledgeBase represents a knowledge base record.
type KnowledgeBase struct {
	ID             string         `json:"id"`
	OrganizationID string         `json:"organization_id"`
	Name           string         `json:"name"`
	Company        string         `json:"company"`
	Status         string         `json:"status"`
	AtlasDraftID   *string        `json:"atlas_draft_id"`
	AtlasLensID    *string        `json:"atlas_lens_id"`
	McpURL         *string        `json:"mcp_url"`
	LensSnapshot   map[string]any `json:"lens_snapshot"`
	LastSyncedAt   *string        `json:"last_synced_at"`
	SyncError      *string        `json:"sync_error"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

// KnowledgeBaseDraftRef is a single ref in a draft selection (names-only
// projection from Atlas). Field names are camelCase on the wire.
type KnowledgeBaseDraftRef struct {
	TypologyID   string   `json:"typologyId"`
	TypologyName string   `json:"typologyName,omitempty"`
	Relevance    string   `json:"relevance"`
	Rationale    string   `json:"rationale,omitempty"`
	TacticIDs    []string `json:"tacticIds,omitempty"`
	TacticNames  []string `json:"tacticNames,omitempty"`
}

// KnowledgeBaseDraftProposal is a staged regeneration awaiting reviewer
// approval.
type KnowledgeBaseDraftProposal struct {
	Refs           []KnowledgeBaseDraftRef `json:"refs"`
	BaseSelection  []KnowledgeBaseDraftRef `json:"baseSelection"`
	Feedback       *string                 `json:"feedback,omitempty"`
	SuggestedName  string                  `json:"suggestedName,omitempty"`
	ProductSummary string                  `json:"productSummary,omitempty"`
	CreatedAt      *string                 `json:"createdAt,omitempty"`
}

// KnowledgeBaseDraft is the projected Atlas draft returned from the poll,
// selection, regenerate, and resolve endpoints.
type KnowledgeBaseDraft struct {
	ID              string                      `json:"id"`
	Status          string                      `json:"status"`
	Error           *string                     `json:"error,omitempty"`
	Company         string                      `json:"company"`
	ProductName     *string                     `json:"productName,omitempty"`
	SuggestedName   string                      `json:"suggestedName"`
	ProductSummary  string                      `json:"productSummary"`
	IterationCount  int                         `json:"iterationCount"`
	Refs            []KnowledgeBaseDraftRef     `json:"refs"`
	PendingProposal *KnowledgeBaseDraftProposal `json:"pendingProposal,omitempty"`
	CreatedAt       *string                     `json:"createdAt,omitempty"`
	UpdatedAt       *string                     `json:"updatedAt,omitempty"`
}

func (k *KnowledgeBaseAPI) orgQuery() map[string]string {
	return map[string]string{"organization_id": k.cfg.OrganizationID}
}

// List returns paginated knowledge bases for the organisation.
func (k *KnowledgeBaseAPI) List(page, pageSize int) (PaginatedResponse[KnowledgeBase], error) {
	return k.ListWithContext(context.Background(), page, pageSize)
}

// ListWithContext returns paginated knowledge bases with a caller-supplied context.
func (k *KnowledgeBaseAPI) ListWithContext(ctx context.Context, page, pageSize int) (PaginatedResponse[KnowledgeBase], error) {
	params := k.orgQuery()
	if page > 0 {
		params["page"] = fmt.Sprintf("%d", page)
	}
	if pageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", pageSize)
	}
	var resp PaginatedResponse[KnowledgeBase]
	if err := k.httpClient.getWithContext(ctx, "/v1/knowledge-base/", params, &resp); err != nil {
		return PaginatedResponse[KnowledgeBase]{}, err
	}
	return resp, nil
}

// Create starts a new knowledge base draft (async generation). Pass empty
// strings for the optional name, productName, and websiteURL.
func (k *KnowledgeBaseAPI) Create(company, brief, name, productName, websiteURL string) (KnowledgeBase, error) {
	return k.CreateWithContext(context.Background(), company, brief, name, productName, websiteURL)
}

// CreateWithContext starts a new knowledge base draft with a caller-supplied context.
func (k *KnowledgeBaseAPI) CreateWithContext(ctx context.Context, company, brief, name, productName, websiteURL string) (KnowledgeBase, error) {
	if company == "" {
		return KnowledgeBase{}, fmt.Errorf("company cannot be empty")
	}
	if brief == "" {
		return KnowledgeBase{}, fmt.Errorf("brief cannot be empty")
	}
	payload := map[string]any{
		"company": company,
		"brief":   brief,
	}
	if name != "" {
		payload["name"] = name
	}
	if productName != "" {
		payload["product_name"] = productName
	}
	if websiteURL != "" {
		payload["website_url"] = websiteURL
	}
	var resp KnowledgeBase
	if err := k.httpClient.postJSONWithContext(ctx, "/v1/knowledge-base/", payload, k.orgQuery(), &resp); err != nil {
		return KnowledgeBase{}, err
	}
	return resp, nil
}

// Retrieve fetches a single knowledge base record.
func (k *KnowledgeBaseAPI) Retrieve(knowledgeBaseID string) (KnowledgeBase, error) {
	return k.RetrieveWithContext(context.Background(), knowledgeBaseID)
}

// RetrieveWithContext fetches a single knowledge base record with a caller-supplied context.
func (k *KnowledgeBaseAPI) RetrieveWithContext(ctx context.Context, knowledgeBaseID string) (KnowledgeBase, error) {
	if knowledgeBaseID == "" {
		return KnowledgeBase{}, fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	var resp KnowledgeBase
	if err := k.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/", knowledgeBaseID), k.orgQuery(), &resp); err != nil {
		return KnowledgeBase{}, err
	}
	return resp, nil
}

// Delete removes a knowledge base and its associated Atlas draft or lens.
func (k *KnowledgeBaseAPI) Delete(knowledgeBaseID string) error {
	return k.DeleteWithContext(context.Background(), knowledgeBaseID)
}

// DeleteWithContext removes a knowledge base with a caller-supplied context.
func (k *KnowledgeBaseAPI) DeleteWithContext(ctx context.Context, knowledgeBaseID string) error {
	if knowledgeBaseID == "" {
		return fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	return k.httpClient.deleteWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/", knowledgeBaseID), k.orgQuery())
}

// Unlink removes the local knowledge base row only, preserving the Atlas lens.
func (k *KnowledgeBaseAPI) Unlink(knowledgeBaseID string) error {
	return k.UnlinkWithContext(context.Background(), knowledgeBaseID)
}

// UnlinkWithContext unlinks a knowledge base with a caller-supplied context.
func (k *KnowledgeBaseAPI) UnlinkWithContext(ctx context.Context, knowledgeBaseID string) error {
	if knowledgeBaseID == "" {
		return fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	return k.httpClient.deleteWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/unlink/", knowledgeBaseID), k.orgQuery())
}

// PollDraft fetches the Atlas draft status. Poll until the status is "ready"
// (or "error").
func (k *KnowledgeBaseAPI) PollDraft(knowledgeBaseID string) (KnowledgeBaseDraft, error) {
	return k.PollDraftWithContext(context.Background(), knowledgeBaseID)
}

// PollDraftWithContext fetches the Atlas draft status with a caller-supplied context.
func (k *KnowledgeBaseAPI) PollDraftWithContext(ctx context.Context, knowledgeBaseID string) (KnowledgeBaseDraft, error) {
	if knowledgeBaseID == "" {
		return KnowledgeBaseDraft{}, fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	var resp KnowledgeBaseDraft
	if err := k.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/draft/", knowledgeBaseID), k.orgQuery(), &resp); err != nil {
		return KnowledgeBaseDraft{}, err
	}
	return resp, nil
}

// PatchSelection persists hand-edits to the draft's typology/tactic selection.
// Pass an empty string for the optional suggestedName.
func (k *KnowledgeBaseAPI) PatchSelection(knowledgeBaseID string, refs []map[string]any, suggestedName string) (KnowledgeBaseDraft, error) {
	return k.PatchSelectionWithContext(context.Background(), knowledgeBaseID, refs, suggestedName)
}

// PatchSelectionWithContext patches the draft selection with a caller-supplied context.
func (k *KnowledgeBaseAPI) PatchSelectionWithContext(ctx context.Context, knowledgeBaseID string, refs []map[string]any, suggestedName string) (KnowledgeBaseDraft, error) {
	if knowledgeBaseID == "" {
		return KnowledgeBaseDraft{}, fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	payload := map[string]any{"refs": refs}
	if suggestedName != "" {
		payload["suggested_name"] = suggestedName
	}
	var resp KnowledgeBaseDraft
	if err := k.httpClient.patchJSONWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/selection/", knowledgeBaseID), payload, k.orgQuery(), &resp); err != nil {
		return KnowledgeBaseDraft{}, err
	}
	return resp, nil
}

// Regenerate kicks off an async regeneration round. Pass an empty string for
// no feedback.
func (k *KnowledgeBaseAPI) Regenerate(knowledgeBaseID, feedback string) (KnowledgeBaseDraft, error) {
	return k.RegenerateWithContext(context.Background(), knowledgeBaseID, feedback)
}

// RegenerateWithContext kicks off a regeneration round with a caller-supplied context.
func (k *KnowledgeBaseAPI) RegenerateWithContext(ctx context.Context, knowledgeBaseID, feedback string) (KnowledgeBaseDraft, error) {
	if knowledgeBaseID == "" {
		return KnowledgeBaseDraft{}, fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	payload := map[string]any{}
	if feedback != "" {
		payload["feedback"] = feedback
	}
	var resp KnowledgeBaseDraft
	if err := k.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/regenerate/", knowledgeBaseID), payload, k.orgQuery(), &resp); err != nil {
		return KnowledgeBaseDraft{}, err
	}
	return resp, nil
}

// Resolve approves or declines a pending regeneration proposal. To apply,
// pass the resolved refs (and optionally suggestedName / acceptSummary); to
// decline, pass discard=true.
func (k *KnowledgeBaseAPI) Resolve(knowledgeBaseID string, refs []map[string]any, suggestedName string, acceptSummary, discard bool) (KnowledgeBaseDraft, error) {
	return k.ResolveWithContext(context.Background(), knowledgeBaseID, refs, suggestedName, acceptSummary, discard)
}

// ResolveWithContext resolves a pending proposal with a caller-supplied context.
func (k *KnowledgeBaseAPI) ResolveWithContext(ctx context.Context, knowledgeBaseID string, refs []map[string]any, suggestedName string, acceptSummary, discard bool) (KnowledgeBaseDraft, error) {
	if knowledgeBaseID == "" {
		return KnowledgeBaseDraft{}, fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	payload := map[string]any{
		"accept_summary": acceptSummary,
		"discard":        discard,
	}
	if refs != nil {
		payload["refs"] = refs
	}
	if suggestedName != "" {
		payload["suggested_name"] = suggestedName
	}
	var resp KnowledgeBaseDraft
	if err := k.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/resolve/", knowledgeBaseID), payload, k.orgQuery(), &resp); err != nil {
		return KnowledgeBaseDraft{}, err
	}
	return resp, nil
}

// Finalize commits the draft into a permanent Atlas lens. Pass an empty
// string to keep the knowledge base's current name. The backend defaults
// mcpEnabled and public to true; pass them explicitly here.
func (k *KnowledgeBaseAPI) Finalize(knowledgeBaseID, name string, mcpEnabled, public bool) (KnowledgeBase, error) {
	return k.FinalizeWithContext(context.Background(), knowledgeBaseID, name, mcpEnabled, public)
}

// FinalizeWithContext finalizes the draft with a caller-supplied context.
func (k *KnowledgeBaseAPI) FinalizeWithContext(ctx context.Context, knowledgeBaseID, name string, mcpEnabled, public bool) (KnowledgeBase, error) {
	if knowledgeBaseID == "" {
		return KnowledgeBase{}, fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	payload := map[string]any{
		"mcp_enabled": mcpEnabled,
		"public":      public,
	}
	if name != "" {
		payload["name"] = name
	}
	var resp KnowledgeBase
	if err := k.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/finalize/", knowledgeBaseID), payload, k.orgQuery(), &resp); err != nil {
		return KnowledgeBase{}, err
	}
	return resp, nil
}

// Sync refreshes the lens snapshot from Atlas (best-effort; always returns
// the record, with any failure described in sync_error).
func (k *KnowledgeBaseAPI) Sync(knowledgeBaseID string) (KnowledgeBase, error) {
	return k.SyncWithContext(context.Background(), knowledgeBaseID)
}

// SyncWithContext refreshes the lens snapshot with a caller-supplied context.
func (k *KnowledgeBaseAPI) SyncWithContext(ctx context.Context, knowledgeBaseID string) (KnowledgeBase, error) {
	if knowledgeBaseID == "" {
		return KnowledgeBase{}, fmt.Errorf("knowledgeBaseID cannot be empty")
	}
	var resp KnowledgeBase
	if err := k.httpClient.postJSONWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/%s/sync/", knowledgeBaseID), nil, k.orgQuery(), &resp); err != nil {
		return KnowledgeBase{}, err
	}
	return resp, nil
}

// Catalog fetches the names-only typology and tactic catalog. The endpoint
// declares no response schema, so the decoded JSON is returned as-is.
func (k *KnowledgeBaseAPI) Catalog() (any, error) {
	return k.CatalogWithContext(context.Background())
}

// CatalogWithContext fetches the catalog with a caller-supplied context.
func (k *KnowledgeBaseAPI) CatalogWithContext(ctx context.Context) (any, error) {
	var resp any
	if err := k.httpClient.getWithContext(ctx, "/v1/knowledge-base/catalog/", k.orgQuery(), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// LensByAtlasId fetches (and best-effort syncs) a lens directly from Atlas by
// its atlas_lens_id. The endpoint declares no response schema, so the decoded
// JSON object is returned as-is.
func (k *KnowledgeBaseAPI) LensByAtlasId(atlasLensID string) (map[string]any, error) {
	return k.LensByAtlasIdWithContext(context.Background(), atlasLensID)
}

// LensByAtlasIdWithContext fetches a lens by Atlas ID with a caller-supplied context.
func (k *KnowledgeBaseAPI) LensByAtlasIdWithContext(ctx context.Context, atlasLensID string) (map[string]any, error) {
	if atlasLensID == "" {
		return nil, fmt.Errorf("atlasLensID cannot be empty")
	}
	var resp map[string]any
	if err := k.httpClient.getWithContext(ctx, fmt.Sprintf("/v1/knowledge-base/lens/%s/", atlasLensID), k.orgQuery(), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ImportLens imports a finalized Atlas lens into roe-main by its
// atlas_lens_id. Idempotent: re-importing an already-tracked lens syncs and
// returns the existing record.
func (k *KnowledgeBaseAPI) ImportLens(atlasLensID string) (KnowledgeBase, error) {
	return k.ImportLensWithContext(context.Background(), atlasLensID)
}

// ImportLensWithContext imports a lens with a caller-supplied context.
func (k *KnowledgeBaseAPI) ImportLensWithContext(ctx context.Context, atlasLensID string) (KnowledgeBase, error) {
	if atlasLensID == "" {
		return KnowledgeBase{}, fmt.Errorf("atlasLensID cannot be empty")
	}
	payload := map[string]any{"atlas_lens_id": atlasLensID}
	var resp KnowledgeBase
	if err := k.httpClient.postJSONWithContext(ctx, "/v1/knowledge-base/import-lens/", payload, k.orgQuery(), &resp); err != nil {
		return KnowledgeBase{}, err
	}
	return resp, nil
}
