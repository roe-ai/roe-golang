package roe

import (
	"net/http"
	"testing"
	"time"
)

func TestAPIErrorTypes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantType   string
	}{
		{
			name:       "BadRequest",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{"detail": "Invalid input"}`),
			wantType:   "*roe.BadRequestError",
		},
		{
			name:       "Unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"message": "Invalid API key"}`),
			wantType:   "*roe.AuthenticationError",
		},
		{
			name:       "PaymentRequired",
			statusCode: http.StatusPaymentRequired,
			body:       []byte(`{"error": "Insufficient credits"}`),
			wantType:   "*roe.InsufficientCreditsError",
		},
		{
			name:       "Forbidden",
			statusCode: http.StatusForbidden,
			body:       []byte(`{"detail": "Access denied"}`),
			wantType:   "*roe.ForbiddenError",
		},
		{
			name:       "NotFound",
			statusCode: http.StatusNotFound,
			body:       []byte(`{"detail": "Agent not found"}`),
			wantType:   "*roe.NotFoundError",
		},
		{
			name:       "TooManyRequests",
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"detail": "Rate limit exceeded"}`),
			wantType:   "*roe.RateLimitError",
		},
		{
			name:       "InternalServerError",
			statusCode: http.StatusInternalServerError,
			body:       []byte(`{"error": "Internal error"}`),
			wantType:   "*roe.ServerError",
		},
		{
			name:       "BadGateway",
			statusCode: http.StatusBadGateway,
			body:       []byte(`Bad gateway`),
			wantType:   "*roe.ServerError",
		},
		{
			name:       "GenericClientError",
			statusCode: 418, // I'm a teapot
			body:       []byte(`{"detail": "I'm a teapot"}`),
			wantType:   "*roe.APIError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			headers.Set("X-Request-ID", "test-request-id")

			err := apiErrorFromResponse(tt.statusCode, tt.body, headers, "X-Request-ID")
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			gotType := getTypeName(err)
			if gotType != tt.wantType {
				t.Errorf("error type = %s, want %s", gotType, tt.wantType)
			}

			// Verify APIError base can be accessed
			apiErr := extractAPIError(err)
			if apiErr == nil {
				t.Error("error should contain *APIError")
				return
			}

			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}

			if apiErr.RequestID != "test-request-id" {
				t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "test-request-id")
			}
		})
	}
}

func TestRateLimitErrorRetryAfter(t *testing.T) {
	tests := []struct {
		name           string
		retryAfterVal  string
		wantNil        bool
		wantApproxSecs int
	}{
		{
			name:           "NumericSeconds",
			retryAfterVal:  "30",
			wantNil:        false,
			wantApproxSecs: 30,
		},
		{
			name:          "Empty",
			retryAfterVal: "",
			wantNil:       true,
		},
		{
			name:          "InvalidValue",
			retryAfterVal: "not-a-number",
			wantNil:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.retryAfterVal != "" {
				headers.Set("Retry-After", tt.retryAfterVal)
			}

			err := apiErrorFromResponse(http.StatusTooManyRequests, []byte(`{}`), headers, "")
			rateLimitErr, ok := err.(*RateLimitError)
			if !ok {
				t.Fatalf("expected *RateLimitError, got %T", err)
			}

			if tt.wantNil {
				if rateLimitErr.RetryAfter != nil {
					t.Errorf("RetryAfter = %v, want nil", *rateLimitErr.RetryAfter)
				}
			} else {
				if rateLimitErr.RetryAfter == nil {
					t.Fatal("RetryAfter = nil, want non-nil")
				}
				gotSecs := int(rateLimitErr.RetryAfter.Seconds())
				if gotSecs != tt.wantApproxSecs {
					t.Errorf("RetryAfter = %d seconds, want %d", gotSecs, tt.wantApproxSecs)
				}
			}
		})
	}
}

func TestExtractErrorDetail(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        []byte
		wantMessage string
	}{
		{
			name:        "DetailField",
			status:      400,
			body:        []byte(`{"detail": "Invalid input"}`),
			wantMessage: "Invalid input",
		},
		{
			name:        "MessageField",
			status:      401,
			body:        []byte(`{"message": "Unauthorized"}`),
			wantMessage: "Unauthorized",
		},
		{
			name:        "ErrorField",
			status:      500,
			body:        []byte(`{"error": "Server error"}`),
			wantMessage: "Server error",
		},
		{
			name:        "PlainTextBody",
			status:      502,
			body:        []byte(`Bad Gateway`),
			wantMessage: "Bad Gateway",
		},
		{
			name:        "EmptyBody",
			status:      503,
			body:        []byte{},
			wantMessage: "HTTP 503",
		},
		{
			name:        "InvalidJSON",
			status:      400,
			body:        []byte(`{invalid json`),
			wantMessage: "{invalid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := extractErrorDetail(tt.status, tt.body)
			if msg != tt.wantMessage {
				t.Errorf("message = %q, want %q", msg, tt.wantMessage)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantNil bool
	}{
		{"NumericSeconds", "60", false},
		{"EmptyValue", "", true},
		{"InvalidString", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.value != "" {
				headers.Set("Retry-After", tt.value)
			}

			result := parseRetryAfter(headers)
			if tt.wantNil && result != nil {
				t.Errorf("expected nil, got %v", *result)
			}
			if !tt.wantNil && result == nil {
				t.Error("expected non-nil, got nil")
			}
		})
	}
}

func TestAPIErrorErrorMethod(t *testing.T) {
	tests := []struct {
		name      string
		err       *APIError
		wantEmpty bool
		contains  string
	}{
		{
			name:      "NilError",
			err:       nil,
			wantEmpty: true,
		},
		{
			name: "WithRequestID",
			err: &APIError{
				StatusCode: 400,
				Message:    "Bad request",
				RequestID:  "req-123",
			},
			contains: "request_id=req-123",
		},
		{
			name: "WithoutRequestID",
			err: &APIError{
				StatusCode: 500,
				Message:    "Server error",
			},
			contains: "roe api error (500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if tt.wantEmpty && result != "" {
				t.Errorf("Error() = %q, want empty", result)
			}
			if tt.contains != "" && !containsString(result, tt.contains) {
				t.Errorf("Error() = %q, want to contain %q", result, tt.contains)
			}
		})
	}
}

func extractAPIError(err error) *APIError {
	switch e := err.(type) {
	case *BadRequestError:
		return e.APIError
	case *AuthenticationError:
		return e.APIError
	case *InsufficientCreditsError:
		return e.APIError
	case *ForbiddenError:
		return e.APIError
	case *NotFoundError:
		return e.APIError
	case *RateLimitError:
		return e.APIError
	case *ServerError:
		return e.APIError
	case *APIError:
		return e
	default:
		return nil
	}
}

func getTypeName(err error) string {
	return "*roe." + getShortTypeName(err)
}

func getShortTypeName(err error) string {
	switch err.(type) {
	case *BadRequestError:
		return "BadRequestError"
	case *AuthenticationError:
		return "AuthenticationError"
	case *InsufficientCreditsError:
		return "InsufficientCreditsError"
	case *ForbiddenError:
		return "ForbiddenError"
	case *NotFoundError:
		return "NotFoundError"
	case *RateLimitError:
		return "RateLimitError"
	case *ServerError:
		return "ServerError"
	case *APIError:
		return "APIError"
	default:
		return "unknown"
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure time import is used
var _ = time.Second
