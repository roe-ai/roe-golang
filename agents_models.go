package roe

import (
	"context"
	"fmt"
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
