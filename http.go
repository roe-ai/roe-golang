package roe

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	mrand "math/rand"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"time"
)

type httpClient struct {
	client    *http.Client
	cfg       Config
	auth      Auth
	logger    Logger
	redactMap map[string]struct{}
}

func newHTTPClient(cfg Config, auth Auth) *httpClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.RetryInitialInterval == 0 {
		cfg.RetryInitialInterval = defaultRetryInitial
	}
	if cfg.RetryMaxInterval == 0 {
		cfg.RetryMaxInterval = defaultRetryMax
	}
	if cfg.RetryMultiplier == 0 {
		cfg.RetryMultiplier = defaultRetryMultiplier
	}
	if cfg.RequestIDHeader == "" {
		cfg.RequestIDHeader = defaultRequestIDHeader
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = defaultMaxIdleConns
	}
	if cfg.MaxIdleConnsPerHost == 0 {
		cfg.MaxIdleConnsPerHost = defaultMaxIdlePerHost
	}
	if cfg.IdleConnTimeout == 0 {
		cfg.IdleConnTimeout = defaultIdleConnTimeout
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
	}
	if cfg.ProxyURL != nil {
		transport.Proxy = http.ProxyURL(cfg.ProxyURL)
	}

	logger := cfg.Logger
	if cfg.Debug && logger == nil {
		logger = log.New(os.Stdout, "roe-sdk ", log.LstdFlags)
	}

	redactions := map[string]struct{}{}
	for _, h := range cfg.RedactHeaders {
		redactions[strings.ToLower(h)] = struct{}{}
	}

	return &httpClient{
		cfg:  cfg,
		auth: auth,
		client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		logger:    logger,
		redactMap: redactions,
	}
}

func (c *httpClient) close() {
	if t, ok := c.client.Transport.(interface{ CloseIdleConnections() }); ok {
		t.CloseIdleConnections()
	}
}

func (c *httpClient) buildURL(path string, query map[string]string) (string, error) {
	base := strings.TrimSuffix(c.cfg.BaseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	full := base + path
	if len(query) == 0 {
		return full, nil
	}
	u, err := url.Parse(full)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// doRequest is a thin shim around doRetried for the legacy hand-written API
// helpers. It builds the request, sends it (with retry), reads the body, and
// converts non-2xx responses to typed errors — preserving the original
// contract of returning ([]byte, error).
func (c *httpClient) doRequest(ctx context.Context, method, path string, headers http.Header, body io.Reader, query map[string]string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	fullURL, err := c.buildURL(path, query)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		if b, ok := body.(*bytes.Buffer); ok {
			bodyReader = bytes.NewReader(b.Bytes())
		} else {
			data, err := io.ReadAll(body)
			if err != nil {
				return nil, fmt.Errorf("read request body: %w", err)
			}
			bodyReader = bytes.NewReader(data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}
	c.applyHeaders(req, headers)

	resp, err := c.doRetried(req)
	if err != nil {
		return nil, err
	}

	respBody, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read response: %w", readErr)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}

	return nil, apiErrorFromResponse(resp.StatusCode, respBody, resp.Header, c.cfg.RequestIDHeader)
}

// doRetried executes a fully-built *http.Request with the SDK's retry policy.
// It is the single retry loop for both the legacy hand-written helpers (via
// doRequest) and the generated raw client (via the retryDoer adapter passed
// to generated.WithHTTPClient). The returned *http.Response always has its
// Body populated with a re-readable buffer; the caller must close Body.
func (c *httpClient) doRetried(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Buffer the body so we can replay it on retries.
	var bodyBytes []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read request body: %w", err)
		}
		bodyBytes = b
	}

	var lastErr error
	maxAttempts := c.cfg.MaxRetries + 1

	// Headers are intentionally NOT re-applied per attempt: req already
	// carries them — set by applyHeaders in doRequest for the legacy path,
	// or by RequestEditorFn (auth) in the retryDoer path before doRetried
	// is invoked. req.Clone(ctx) preserves the header map, so each retry
	// inherits the same auth/static headers without redundant work.
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		attemptReq := req.Clone(ctx)
		if bodyBytes != nil {
			attemptReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			attemptReq.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(bodyBytes)), nil
			}
			attemptReq.ContentLength = int64(len(bodyBytes))
		} else {
			attemptReq.Body = nil
			attemptReq.GetBody = nil
			attemptReq.ContentLength = 0
		}

		c.attachRequestID(attemptReq)
		c.runRequestHooks(attemptReq)
		c.logRequest(attemptReq, attempt)

		start := time.Now()
		resp, err := c.client.Do(attemptReq)
		duration := time.Since(start)

		if err != nil {
			if !c.shouldRetry(nil, err, attempt) {
				return nil, err
			}
			lastErr = err
			c.logf("retrying after error (attempt %d/%d): %v", attempt+1, maxAttempts, err)
			if sleepErr := c.sleepWithContext(ctx, c.backoffDuration(attempt)); sleepErr != nil {
				return nil, sleepErr
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read response: %w", readErr)
		}

		c.logResponse(attemptReq, resp, respBody, duration)
		c.runResponseHooks(resp, respBody)

		if c.shouldRetry(resp, nil, attempt) {
			lastErr = apiErrorFromResponse(resp.StatusCode, respBody, resp.Header, c.cfg.RequestIDHeader)
			c.logf("retrying after status %d (attempt %d/%d)", resp.StatusCode, attempt+1, maxAttempts)
			if sleepErr := c.sleepWithContext(ctx, c.retryDelay(resp, attempt)); sleepErr != nil {
				return nil, sleepErr
			}
			continue
		}

		// Restore the body so the caller can read it.
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		return resp, nil
	}

	return nil, lastErr
}

func (c *httpClient) logf(format string, args ...any) {
	if c.logger == nil || !c.cfg.Debug {
		return
	}
	c.logger.Printf(format, args...)
}

func (c *httpClient) logRequest(req *http.Request, attempt int) {
	if c.logger == nil || !c.cfg.Debug {
		return
	}
	c.logger.Printf("[request] %s %s attempt=%d headers=%v", req.Method, req.URL.String(), attempt+1, c.redactedHeaders(req.Header))
}

func (c *httpClient) logResponse(req *http.Request, resp *http.Response, body []byte, duration time.Duration) {
	if c.logger == nil || !c.cfg.Debug {
		return
	}
	requestID := resp.Header.Get(c.cfg.RequestIDHeader)
	bodyPreview := string(body)
	if len(bodyPreview) > 512 {
		bodyPreview = bodyPreview[:512] + "…"
	}
	c.logger.Printf("[response] %s %s status=%d duration=%s request_id=%s body=%s", req.Method, req.URL.String(), resp.StatusCode, duration, requestID, bodyPreview)
}

func (c *httpClient) redactedHeaders(h http.Header) http.Header {
	if len(c.redactMap) == 0 {
		return h
	}
	cloned := cloneHeaders(h)
	for k := range cloned {
		if _, ok := c.redactMap[strings.ToLower(k)]; ok {
			cloned.Set(k, "[redacted]")
		}
	}
	return cloned
}

func (c *httpClient) applyHeaders(req *http.Request, headers http.Header) {
	for k, vals := range c.auth.Headers() {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	for k, vals := range c.cfg.ExtraHeaders {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	for k, vals := range headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
}

func (c *httpClient) attachRequestID(req *http.Request) {
	if c.cfg.RequestIDHeader == "" {
		return
	}
	if req.Header.Get(c.cfg.RequestIDHeader) != "" {
		return
	}
	switch {
	case c.cfg.DefaultRequestID != "":
		req.Header.Set(c.cfg.RequestIDHeader, c.cfg.DefaultRequestID)
	case c.cfg.AutoRequestID:
		req.Header.Set(c.cfg.RequestIDHeader, c.generateRequestID())
	}
}

func (c *httpClient) generateRequestID() string {
	buf := make([]byte, 12)
	if _, err := crand.Read(buf); err == nil {
		return "roe-" + hex.EncodeToString(buf)
	}
	return fmt.Sprintf("roe-%d", time.Now().UnixNano())
}

func (c *httpClient) runRequestHooks(req *http.Request) {
	for i, hook := range c.cfg.BeforeRequest {
		func() {
			defer func() {
				if r := recover(); r != nil {
					c.logf("request hook[%d] panic: %v", i, r)
				}
			}()
			hook(req)
		}()
	}
}

func (c *httpClient) runResponseHooks(resp *http.Response, body []byte) {
	for i, hook := range c.cfg.AfterResponse {
		func() {
			defer func() {
				if r := recover(); r != nil {
					c.logf("response hook[%d] panic: %v", i, r)
				}
			}()
			hook(resp, body)
		}()
	}
}

func (c *httpClient) shouldRetry(resp *http.Response, err error, attempt int) bool {
	if attempt >= c.cfg.MaxRetries {
		return false
	}
	if err != nil {
		// Don't retry on context cancellation/timeout
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		// Only retry on temporary network errors
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return true
		}
		// Retry on connection errors (DNS, connection refused, etc.)
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return true
		}
		// Don't retry other errors (permanent failures)
		return false
	}
	if resp == nil {
		return false
	}
	if resp.StatusCode >= 500 {
		return true
	}
	switch resp.StatusCode {
	case http.StatusRequestTimeout, http.StatusTooManyRequests:
		return true
	default:
		return false
	}
}

func (c *httpClient) backoffDuration(attempt int) time.Duration {
	factor := math.Pow(c.cfg.RetryMultiplier, float64(attempt))
	delay := time.Duration(float64(c.cfg.RetryInitialInterval) * factor)
	if delay > c.cfg.RetryMaxInterval {
		delay = c.cfg.RetryMaxInterval
	}
	if c.cfg.RetryJitter > 0 {
		jitterFactor := 1 + (mrand.Float64()*2-1)*c.cfg.RetryJitter
		delay = time.Duration(float64(delay) * jitterFactor)
	}
	if delay < time.Millisecond {
		return time.Millisecond
	}
	return delay
}

func (c *httpClient) retryDelay(resp *http.Response, attempt int) time.Duration {
	delay := c.backoffDuration(attempt)
	if resp == nil {
		return delay
	}
	retryAfter := parseRetryAfter(resp.Header)
	if retryAfter == nil || *retryAfter <= 0 {
		return delay
	}
	if *retryAfter > delay {
		return *retryAfter
	}
	return delay
}

func (c *httpClient) sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type preparedFile struct {
	FieldName string
	File      FileUpload
}

// dynamicInputsRequest builds the body and Content-Type for a dynamic-input
// agent run request. It does NOT send the request — the wrapper layer feeds
// the result into the generated client's *WithBodyWithResponse helper, which
// owns the URL/path-param/query construction.
//
// Detection rules match the legacy postDynamicInputsWithContext helper:
//   - FileUpload / *FileUpload (URL-only forms collapse to a form field)
//   - *bytes.Buffer / *bytes.Reader / []byte / io.Reader → multipart file part
//   - string: UUID → form, file path → multipart file, looks-like-path-but-
//     missing → error, otherwise → form
//   - fmt.Stringer → form
//   - everything else → form via fmt.Sprintf("%v")
//
// When metadata is non-nil it is JSON-encoded and added as a "metadata" form
// field; passing a "metadata" key in inputs alongside a non-nil metadata
// argument is rejected.
func (c *httpClient) dynamicInputsRequest(inputs map[string]any, metadata map[string]any) (io.Reader, string, error) {
	if metadata != nil {
		if _, exists := inputs["metadata"]; exists {
			return nil, "", fmt.Errorf("inputs must not contain key \"metadata\" when metadata parameter is set")
		}
	}

	form := url.Values{}
	var files []preparedFile

	for key, val := range inputs {
		switch v := val.(type) {
		case FileUpload:
			if v.isURL() && v.Path == "" && v.Reader == nil {
				form.Set(key, v.URL)
			} else {
				files = append(files, preparedFile{FieldName: key, File: v})
			}
		case *FileUpload:
			if v != nil {
				if v.isURL() && v.Path == "" && v.Reader == nil {
					form.Set(key, v.URL)
				} else {
					files = append(files, preparedFile{FieldName: key, File: *v})
				}
			}
		case *bytes.Buffer:
			files = append(files, preparedFile{FieldName: key, File: FileUpload{Reader: bytes.NewReader(v.Bytes()), Filename: key}})
		case *bytes.Reader:
			files = append(files, preparedFile{FieldName: key, File: FileUpload{Reader: v, Filename: key}})
		case []byte:
			files = append(files, preparedFile{FieldName: key, File: FileUpload{Reader: bytes.NewReader(v), Filename: key}})
		case io.Reader:
			files = append(files, preparedFile{FieldName: key, File: FileUpload{Reader: v, Filename: key}})
		case string:
			switch {
			case isUUIDString(v):
				form.Set(key, v)
			case isFilePath(v):
				files = append(files, preparedFile{FieldName: key, File: FileUpload{Path: v}})
			case looksLikePath(v) && !isHTTPURL(v):
				return nil, "", fmt.Errorf("input %s references a file that was not found: %s", key, v)
			default:
				form.Set(key, v)
			}
		case fmt.Stringer:
			form.Set(key, v.String())
		default:
			form.Set(key, fmt.Sprintf("%v", v))
		}
	}

	// Serialize metadata as JSON string form field after inputs are processed.
	if metadata != nil {
		metaJSON, err := json.Marshal(metadata)
		if err != nil {
			return nil, "", fmt.Errorf("marshal metadata: %w", err)
		}
		form.Set("metadata", string(metaJSON))
	}

	if len(files) == 0 {
		return strings.NewReader(form.Encode()), "application/x-www-form-urlencoded", nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, values := range form {
		for _, v := range values {
			_ = writer.WriteField(key, v)
		}
	}

	// Track opened readers for cleanup on error.
	var openedReaders []io.ReadCloser
	closeAllReaders := func() {
		for _, r := range openedReaders {
			r.Close()
		}
	}

	for _, f := range files {
		fileReader, filename, mimeType, err := c.prepareMultipartFile(f.File)
		if err != nil {
			closeAllReaders()
			writer.Close()
			return nil, "", err
		}
		openedReaders = append(openedReaders, fileReader)

		h := make(textproto.MIMEHeader)
		// Use mime.FormatMediaType to properly escape filename (handles quotes, newlines, etc.)
		contentDisp := mime.FormatMediaType("form-data", map[string]string{
			"name":     f.FieldName,
			"filename": filename,
		})
		h.Set("Content-Disposition", contentDisp)
		h.Set("Content-Type", mimeType)

		part, err := writer.CreatePart(h)
		if err != nil {
			closeAllReaders()
			writer.Close()
			return nil, "", err
		}
		if _, err := io.Copy(part, fileReader); err != nil {
			closeAllReaders()
			writer.Close()
			return nil, "", err
		}
		fileReader.Close()
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func (c *httpClient) prepareMultipartFile(file FileUpload) (io.ReadCloser, string, string, error) {
	rc, err := file.open()
	if err != nil {
		return nil, "", "", err
	}

	filename := file.filename()
	mimeType := file.mimeType()

	rcWithMime, detected, err := detectMimeType(rc, filename, mimeType)
	if err != nil {
		return nil, "", "", err
	}
	if detected != "" {
		mimeType = detected
	}

	return rcWithMime, filename, mimeType, nil
}

func detectMimeType(rc io.ReadCloser, filename, fallback string) (io.ReadCloser, string, error) {
	if seeker, ok := rc.(io.ReadSeeker); ok {
		buf := make([]byte, 512)
		n, _ := seeker.Read(buf)
		_, _ = seeker.Seek(0, io.SeekStart)
		if n > 0 {
			detected := http.DetectContentType(buf[:n])
			return rc, detected, nil
		}
		return rc, fallback, nil
	}

	buf, err := io.ReadAll(io.LimitReader(rc, 1024))
	if err != nil {
		rc.Close()
		return nil, "", err
	}

	detected := http.DetectContentType(buf)
	combined := io.NopCloser(io.MultiReader(bytes.NewReader(buf), rc))
	if detected != "" {
		return combined, detected, nil
	}
	return combined, fallback, nil
}
