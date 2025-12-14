package roe

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Logger is the minimal logging interface supported by the SDK.
type Logger interface {
	Printf(format string, v ...any)
}

// RequestHook allows callers to inspect or mutate requests before they are sent.
type RequestHook func(*http.Request)

// ResponseHook allows callers to inspect responses (raw bytes included).
type ResponseHook func(*http.Response, []byte)

// Config holds SDK configuration.
type Config struct {
	APIKey         string
	OrganizationID string
	BaseURL        string
	Timeout        time.Duration
	MaxRetries     int

	Debug bool

	ExtraHeaders http.Header
	ProxyURL     *url.URL

	RequestIDHeader  string
	DefaultRequestID string
	AutoRequestID    bool

	RetryInitialInterval time.Duration
	RetryMaxInterval     time.Duration
	RetryMultiplier      float64
	RetryJitter          float64

	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration

	Logger        Logger
	RedactHeaders []string

	BeforeRequest []RequestHook
	AfterResponse []ResponseHook
}

// ConfigParams provides optional overrides for building a Config.
type ConfigParams struct {
	APIKey          string
	OrganizationID  string
	BaseURL         string
	Timeout         time.Duration
	TimeoutSeconds  float64
	MaxRetries      int
	Debug           *bool
	ExtraHeaders    http.Header
	ProxyURL        string
	RequestID       string
	AutoRequestID   *bool
	RequestIDHeader string

	RetryInitialInterval time.Duration
	RetryMaxInterval     time.Duration
	RetryMultiplier      float64
	RetryJitter          float64

	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration

	Logger        Logger
	RedactHeaders []string

	BeforeRequest []RequestHook
	AfterResponse []ResponseHook
}

const (
	defaultBaseURL         = "https://api.roe-ai.com"
	defaultTimeout         = 60 * time.Second
	defaultMaxRetries      = 3
	defaultRetryInitial    = 200 * time.Millisecond
	defaultRetryMax        = 2 * time.Second
	defaultRetryMultiplier = 2.0
	defaultRetryJitter     = 0.2
	defaultMaxIdleConns    = 100
	defaultMaxIdlePerHost  = 10
	defaultIdleConnTimeout = 90 * time.Second
	defaultRequestIDHeader = "X-Request-ID"
)

// LoadConfig builds a Config from parameters or environment variables.
// Environment fallbacks:
//
//	ROE_API_KEY, ROE_ORGANIZATION_ID, ROE_BASE_URL, ROE_TIMEOUT, ROE_MAX_RETRIES,
//	ROE_DEBUG, ROE_PROXY, ROE_EXTRA_HEADERS, ROE_REQUEST_ID, ROE_AUTO_REQUEST_ID,
//	ROE_REQUEST_ID_HEADER, ROE_RETRY_INITIAL_MS, ROE_RETRY_MAX_MS,
//	ROE_RETRY_MULTIPLIER, ROE_RETRY_JITTER, ROE_MAX_IDLE_CONNS,
//	ROE_MAX_IDLE_CONNS_PER_HOST, ROE_IDLE_CONN_TIMEOUT.
func LoadConfig(apiKey, orgID, baseURL string, timeoutSeconds float64, maxRetries int) (Config, error) {
	return LoadConfigWithParams(ConfigParams{
		APIKey:         apiKey,
		OrganizationID: orgID,
		BaseURL:        baseURL,
		TimeoutSeconds: timeoutSeconds,
		MaxRetries:     maxRetries,
	})
}

// LoadConfigWithParams is an extended constructor that accepts structured options.
func LoadConfigWithParams(params ConfigParams) (Config, error) {
	envIdleTimeout, err := parseEnvDuration("ROE_IDLE_CONN_TIMEOUT", time.Second)
	if err != nil {
		return Config{}, err
	}

	envMaxRetries, envMaxRetriesSet, err := parseEnvInt("ROE_MAX_RETRIES")
	if err != nil {
		return Config{}, err
	}
	envMaxIdleConns, envMaxIdleConnsSet, err := parseEnvInt("ROE_MAX_IDLE_CONNS")
	if err != nil {
		return Config{}, err
	}
	envMaxIdlePerHost, envMaxIdlePerHostSet, err := parseEnvInt("ROE_MAX_IDLE_CONNS_PER_HOST")
	if err != nil {
		return Config{}, err
	}

	maxRetries := defaultMaxRetries
	if envMaxRetriesSet {
		maxRetries = envMaxRetries
	}
	if params.MaxRetries != 0 {
		maxRetries = params.MaxRetries
	}

	maxIdleConns := defaultMaxIdleConns
	if envMaxIdleConnsSet {
		maxIdleConns = envMaxIdleConns
	}
	if params.MaxIdleConns != 0 {
		maxIdleConns = params.MaxIdleConns
	}

	maxIdlePerHost := defaultMaxIdlePerHost
	if envMaxIdlePerHostSet {
		maxIdlePerHost = envMaxIdlePerHost
	}
	if params.MaxIdleConnsPerHost != 0 {
		maxIdlePerHost = params.MaxIdleConnsPerHost
	}

	cfg := Config{
		APIKey:               firstNonEmpty(params.APIKey, os.Getenv("ROE_API_KEY")),
		OrganizationID:       firstNonEmpty(params.OrganizationID, os.Getenv("ROE_ORGANIZATION_ID")),
		BaseURL:              firstNonEmpty(params.BaseURL, os.Getenv("ROE_BASE_URL"), defaultBaseURL),
		MaxRetries:           maxRetries,
		ExtraHeaders:         cloneHeaders(params.ExtraHeaders),
		RequestIDHeader:      firstNonEmpty(params.RequestIDHeader, os.Getenv("ROE_REQUEST_ID_HEADER"), defaultRequestIDHeader),
		DefaultRequestID:     firstNonEmpty(params.RequestID, os.Getenv("ROE_REQUEST_ID")),
		RetryInitialInterval: defaultRetryInitial,
		RetryMaxInterval:     defaultRetryMax,
		RetryMultiplier:      defaultRetryMultiplier,
		RetryJitter:          defaultRetryJitter,
		MaxIdleConns:         maxIdleConns,
		MaxIdleConnsPerHost:  maxIdlePerHost,
		IdleConnTimeout:      firstNonZeroDuration(params.IdleConnTimeout, envIdleTimeout, defaultIdleConnTimeout),
		Logger:               params.Logger,
		RedactHeaders:        params.RedactHeaders,
		BeforeRequest:        params.BeforeRequest,
		AfterResponse:        params.AfterResponse,
		AutoRequestID:        true,
	}

	if cfg.ExtraHeaders == nil {
		cfg.ExtraHeaders = http.Header{}
	}
	if cfg.RedactHeaders == nil {
		cfg.RedactHeaders = []string{"Authorization", "X-API-Key", "X-Request-ID"}
	}

	if params.Debug != nil {
		cfg.Debug = *params.Debug
	} else if env := os.Getenv("ROE_DEBUG"); env != "" {
		val, err := strconv.ParseBool(env)
		if err != nil {
			return Config{}, fmt.Errorf("parse ROE_DEBUG: %w", err)
		}
		cfg.Debug = val
	}

	if params.Timeout > 0 {
		cfg.Timeout = params.Timeout
	} else if params.TimeoutSeconds > 0 {
		cfg.Timeout = time.Duration(params.TimeoutSeconds * float64(time.Second))
	} else if envTimeout, err := parseEnvDuration("ROE_TIMEOUT", time.Second); err != nil {
		return Config{}, err
	} else if envTimeout > 0 {
		cfg.Timeout = envTimeout
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.Timeout < 0 {
		return Config{}, fmt.Errorf("timeout must be non-negative")
	}

	if env := os.Getenv("ROE_EXTRA_HEADERS"); env != "" {
		envHeaders, err := parseHeadersEnv(env)
		if err != nil {
			return Config{}, err
		}
		for k, vals := range envHeaders {
			for _, v := range vals {
				cfg.ExtraHeaders.Add(k, v)
			}
		}
	}

	proxyURL := params.ProxyURL
	if proxyURL == "" {
		proxyURL = os.Getenv("ROE_PROXY")
	}
	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return Config{}, fmt.Errorf("parse ROE_PROXY: %w", err)
		}
		cfg.ProxyURL = parsed
	}

	if params.AutoRequestID != nil {
		cfg.AutoRequestID = *params.AutoRequestID
	} else if env := os.Getenv("ROE_AUTO_REQUEST_ID"); env != "" {
		val, err := strconv.ParseBool(env)
		if err != nil {
			return Config{}, fmt.Errorf("parse ROE_AUTO_REQUEST_ID: %w", err)
		}
		cfg.AutoRequestID = val
	}

	if val, err := parseEnvDuration("ROE_RETRY_INITIAL_MS", time.Millisecond); err != nil {
		return Config{}, err
	} else if val > 0 {
		cfg.RetryInitialInterval = val
	}
	if val, err := parseEnvDuration("ROE_RETRY_MAX_MS", time.Millisecond); err != nil {
		return Config{}, err
	} else if val > 0 {
		cfg.RetryMaxInterval = val
	}
	if valStr := os.Getenv("ROE_RETRY_MULTIPLIER"); valStr != "" {
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return Config{}, fmt.Errorf("parse ROE_RETRY_MULTIPLIER: %w", err)
		}
		cfg.RetryMultiplier = val
	}
	if valStr := os.Getenv("ROE_RETRY_JITTER"); valStr != "" {
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return Config{}, fmt.Errorf("parse ROE_RETRY_JITTER: %w", err)
		}
		cfg.RetryJitter = val
	}

	if cfg.APIKey == "" {
		return Config{}, ErrMissingAPIKey
	}
	if cfg.OrganizationID == "" {
		return Config{}, ErrMissingOrganizationID
	}
	if cfg.MaxRetries < 0 {
		return Config{}, fmt.Errorf("max retries must be >= 0")
	}
	if cfg.MaxIdleConns < 0 {
		return Config{}, fmt.Errorf("max idle conns must be >= 0")
	}
	if cfg.MaxIdleConnsPerHost < 0 {
		return Config{}, fmt.Errorf("max idle conns per host must be >= 0")
	}
	if cfg.IdleConnTimeout < 0 {
		return Config{}, fmt.Errorf("idle connection timeout must be non-negative")
	}
	if cfg.RetryInitialInterval <= 0 || cfg.RetryMaxInterval <= 0 {
		return Config{}, fmt.Errorf("retry intervals must be positive")
	}
	if cfg.RetryMultiplier < 1 {
		return Config{}, fmt.Errorf("retry multiplier must be >= 1")
	}
	if cfg.RetryJitter < 0 || cfg.RetryJitter > 1 {
		return Config{}, fmt.Errorf("retry jitter must be between 0 and 1")
	}

	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}


func firstNonZeroDuration(values ...time.Duration) time.Duration {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}

func parseEnvInt(env string) (int, bool, error) {
	val, ok := os.LookupEnv(env)
	if !ok || val == "" {
		return 0, false, nil
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, true, fmt.Errorf("parse %s: %w", env, err)
	}
	return parsed, true, nil
}

func parseEnvDuration(env string, numericUnit time.Duration) (time.Duration, error) {
	val := os.Getenv(env)
	if val == "" {
		return 0, nil
	}
	if duration, err := time.ParseDuration(val); err == nil {
		return duration, nil
	}
	seconds, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", env, err)
	}
	return time.Duration(seconds * float64(numericUnit)), nil
}

func parseHeadersEnv(val string) (http.Header, error) {
	headers := http.Header{}
	if val == "" {
		return headers, nil
	}
	for _, entry := range strings.FieldsFunc(val, func(r rune) bool { return r == ';' || r == ',' || r == '\n' }) {
		if entry == "" {
			continue
		}
		sep := ":"
		if strings.Contains(entry, "=") {
			sep = "="
		}
		parts := strings.SplitN(entry, sep, 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header entry %q", entry)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			return nil, fmt.Errorf("invalid header entry %q", entry)
		}
		headers.Add(key, value)
	}
	return headers, nil
}

func cloneHeaders(h http.Header) http.Header {
	if h == nil {
		return http.Header{}
	}
	clone := http.Header{}
	for k, vals := range h {
		clone[k] = append([]string(nil), vals...)
	}
	return clone
}
