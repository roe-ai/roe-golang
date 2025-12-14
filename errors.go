package roe

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrMissingAPIKey         = errors.New("API key is required. Provide it or set ROE_API_KEY")
	ErrMissingOrganizationID = errors.New("Organization ID is required. Provide it or set ROE_ORGANIZATION_ID")
)

// APIError represents an error returned by the Roe API.
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
	RequestID  string
	Details    map[string]any
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.RequestID != "" {
		return fmt.Sprintf("roe api error (%d): %s (request_id=%s)", e.StatusCode, e.Message, e.RequestID)
	}
	return fmt.Sprintf("roe api error (%d): %s", e.StatusCode, e.Message)
}

type BadRequestError struct{ *APIError }
type AuthenticationError struct{ *APIError }
type InsufficientCreditsError struct{ *APIError }
type ForbiddenError struct{ *APIError }
type NotFoundError struct{ *APIError }
type RateLimitError struct {
	*APIError
	RetryAfter *time.Duration
}
type ServerError struct{ *APIError }

// apiErrorFromResponse maps an HTTP status code and optional JSON body to a typed error.
func apiErrorFromResponse(status int, body []byte, headers http.Header, requestIDHeader string) error {
	message, details := extractErrorDetail(status, body)
	requestID := ""
	if headers != nil && requestIDHeader != "" {
		requestID = headers.Get(requestIDHeader)
	}

	base := &APIError{
		StatusCode: status,
		Message:    message,
		Body:       body,
		RequestID:  requestID,
		Details:    details,
	}

	switch status {
	case http.StatusBadRequest:
		return &BadRequestError{APIError: base}
	case http.StatusUnauthorized:
		return &AuthenticationError{APIError: base}
	case http.StatusPaymentRequired:
		return &InsufficientCreditsError{APIError: base}
	case http.StatusForbidden:
		return &ForbiddenError{APIError: base}
	case http.StatusNotFound:
		return &NotFoundError{APIError: base}
	case http.StatusTooManyRequests:
		return &RateLimitError{APIError: base, RetryAfter: parseRetryAfter(headers)}
	default:
		if status >= 500 {
			return &ServerError{APIError: base}
		}
		return base
	}
}

func extractErrorDetail(status int, body []byte) (string, map[string]any) {
	details := map[string]any{}
	if len(body) == 0 {
		return fmt.Sprintf("HTTP %d", status), details
	}
	raw := strings.TrimSpace(string(body))

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err == nil {
		details = parsed
		if msg := findDetailString(parsed); msg != "" {
			return msg, details
		}
	}
	if raw != "" {
		return raw, details
	}
	return fmt.Sprintf("HTTP %d", status), details
}

func findDetailString(parsed map[string]any) string {
	for _, key := range []string{"detail", "message", "error"} {
		if v, ok := parsed[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func parseRetryAfter(headers http.Header) *time.Duration {
	if headers == nil {
		return nil
	}
	val := headers.Get("Retry-After")
	if val == "" {
		return nil
	}
	if seconds, err := strconv.Atoi(val); err == nil {
		d := time.Duration(seconds) * time.Second
		return &d
	}
	if t, err := http.ParseTime(val); err == nil {
		d := time.Until(t)
		return &d
	}
	return nil
}
