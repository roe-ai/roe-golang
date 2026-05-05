package roe

import "net/http"

// retryDoer adapts the SDK's *httpClient retry loop to the
// generated.HttpRequestDoer interface so the generated raw client gets the
// same retry/hooks/redaction/request-ID behavior as the legacy hand-written
// helpers in http.go.
type retryDoer struct {
	c *httpClient
}

func (d *retryDoer) Do(req *http.Request) (*http.Response, error) {
	return d.c.doRetried(req)
}
