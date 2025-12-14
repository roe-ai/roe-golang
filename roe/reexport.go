package roe

import root "github.com/roe-ai/roe-golang"

type (
	// Core client/config.
	RoeClient    = root.RoeClient
	Config       = root.Config
	ConfigParams = root.ConfigParams
	Logger       = root.Logger

	RequestHook  = root.RequestHook
	ResponseHook = root.ResponseHook

	// API surfaces.
	Auth               = root.Auth
	AgentsAPI          = root.AgentsAPI
	AgentVersionsAPI   = root.AgentVersionsAPI
	AgentJobsAPI       = root.AgentJobsAPI
	ListVersionsParams = root.ListVersionsParams

	// Models/results.
	AgentInputDefinition = root.AgentInputDefinition
	UserInfo             = root.UserInfo
	BaseAgent            = root.BaseAgent
	AgentVersion         = root.AgentVersion

	Job      = root.Job
	JobBatch = root.JobBatch

	JobStatus = root.JobStatus

	AgentDatum     = root.AgentDatum
	AgentJobStatus = root.AgentJobStatus
	Reference                = root.Reference
	AgentJobResult           = root.AgentJobResult
	AgentJobStatusBatch      = root.AgentJobStatusBatch
	AgentJobResultBatch      = root.AgentJobResultBatch
	JobDataDeleteResponse    = root.JobDataDeleteResponse

	// File uploads.
	FileUpload = root.FileUpload

	// Errors.
	APIError                 = root.APIError
	BadRequestError          = root.BadRequestError
	AuthenticationError      = root.AuthenticationError
	InsufficientCreditsError = root.InsufficientCreditsError
	ForbiddenError           = root.ForbiddenError
	NotFoundError            = root.NotFoundError
	RateLimitError           = root.RateLimitError
	ServerError              = root.ServerError
)

const (
	JobPending   = root.JobPending
	JobStarted   = root.JobStarted
	JobRetry     = root.JobRetry
	JobSuccess   = root.JobSuccess
	JobFailure   = root.JobFailure
	JobCancelled = root.JobCancelled
	JobCached    = root.JobCached
)

var (
	ErrMissingAPIKey         = root.ErrMissingAPIKey
	ErrMissingOrganizationID = root.ErrMissingOrganizationID
)

func NewClient(apiKey, organizationID, baseURL string, timeoutSeconds float64, maxRetries int) (*RoeClient, error) {
	return root.NewClient(apiKey, organizationID, baseURL, timeoutSeconds, maxRetries)
}

func NewClientWithParams(params ConfigParams) (*RoeClient, error) {
	return root.NewClientWithParams(params)
}

func NewClientWithConfig(cfg Config) (*RoeClient, error) {
	return root.NewClientWithConfig(cfg)
}

func LoadConfig(apiKey, orgID, baseURL string, timeoutSeconds float64, maxRetries int) (Config, error) {
	return root.LoadConfig(apiKey, orgID, baseURL, timeoutSeconds, maxRetries)
}

func LoadConfigWithParams(params ConfigParams) (Config, error) {
	return root.LoadConfigWithParams(params)
}
