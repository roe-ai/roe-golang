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
	"mime/multipart"
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

func (c *httpClient) doRequest(ctx context.Context, method, path string, headers http.Header, body io.Reader, query map[string]string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	fullURL, err := c.buildURL(path, query)
	if err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if body != nil {
		if b, ok := body.(*bytes.Buffer); ok {
			bodyBytes = b.Bytes()
		} else {
			bodyBytes, err = io.ReadAll(body)
			if err != nil {
				return nil, fmt.Errorf("read request body: %w", err)
			}
		}
	}

	var lastErr error
	maxAttempts := c.cfg.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return nil, err
		}

		c.applyHeaders(req, headers)
		c.attachRequestID(req)
		c.runRequestHooks(req)
		c.logRequest(req, attempt)

		start := time.Now()
		resp, err := c.client.Do(req)
		duration := time.Since(start)

		if err != nil {
			if !c.shouldRetry(nil, err, attempt) {
				return nil, err
			}
			lastErr = err
			c.logf("retrying after error (attempt %d/%d): %v", attempt+1, maxAttempts, err)
			if err := c.sleepWithContext(ctx, c.backoffDuration(attempt)); err != nil {
				return nil, err
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read response: %w", readErr)
		}

		c.logResponse(req, resp, respBody, duration)
		c.runResponseHooks(resp, respBody)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		apiErr := apiErrorFromResponse(resp.StatusCode, respBody, resp.Header, c.cfg.RequestIDHeader)
		lastErr = apiErr

		if c.shouldRetry(resp, nil, attempt) {
			c.logf("retrying after status %d (attempt %d/%d)", resp.StatusCode, attempt+1, maxAttempts)
			if err := c.sleepWithContext(ctx, c.retryDelay(resp, attempt)); err != nil {
				return nil, err
			}
			continue
		}

		return nil, apiErr
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		return true
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

func (c *httpClient) get(path string, query map[string]string, out any) error {
	return c.getWithContext(context.Background(), path, query, out)
}

func (c *httpClient) getWithContext(ctx context.Context, path string, query map[string]string, out any) error {
	data, err := c.doRequest(ctx, http.MethodGet, path, http.Header{}, nil, query)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (c *httpClient) getBytes(path string, query map[string]string) ([]byte, error) {
	return c.getBytesWithContext(context.Background(), path, query)
}

func (c *httpClient) getBytesWithContext(ctx context.Context, path string, query map[string]string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, http.Header{}, nil, query)
}

func (c *httpClient) delete(path string, query map[string]string) error {
	return c.deleteWithContext(context.Background(), path, query)
}

func (c *httpClient) deleteWithContext(ctx context.Context, path string, query map[string]string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, path, http.Header{}, nil, query)
	return err
}

func (c *httpClient) postJSON(path string, payload any, query map[string]string, out any) error {
	return c.postJSONWithContext(context.Background(), path, payload, query, out)
}

func (c *httpClient) postJSONWithContext(ctx context.Context, path string, payload any, query map[string]string, out any) error {
	buf := &bytes.Buffer{}
	if payload != nil {
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
	}
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	data, err := c.doRequest(ctx, http.MethodPost, path, headers, buf, query)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (c *httpClient) putJSON(path string, payload any, query map[string]string, out any) error {
	return c.putJSONWithContext(context.Background(), path, payload, query, out)
}

func (c *httpClient) putJSONWithContext(ctx context.Context, path string, payload any, query map[string]string, out any) error {
	buf := &bytes.Buffer{}
	if payload != nil {
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
	}
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	data, err := c.doRequest(ctx, http.MethodPut, path, headers, buf, query)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

type preparedFile struct {
	FieldName string
	File      FileUpload
}

func (c *httpClient) postDynamicInputs(path string, inputs map[string]any, query map[string]string, out any) error {
	return c.postDynamicInputsWithContext(context.Background(), path, inputs, query, out)
}

func (c *httpClient) postDynamicInputsWithContext(ctx context.Context, path string, inputs map[string]any, query map[string]string, out any) error {
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
		case io.Reader:
			files = append(files, preparedFile{FieldName: key, File: FileUpload{Reader: v, Filename: key}})
		case []byte:
			files = append(files, preparedFile{FieldName: key, File: FileUpload{Reader: bytes.NewReader(v), Filename: key}})
		case *bytes.Buffer:
			files = append(files, preparedFile{FieldName: key, File: FileUpload{Reader: bytes.NewReader(v.Bytes()), Filename: key}})
		case *bytes.Reader:
			files = append(files, preparedFile{FieldName: key, File: FileUpload{Reader: v, Filename: key}})
		case string:
			switch {
			case isUUIDString(v):
				form.Set(key, v)
			case isFilePath(v):
				files = append(files, preparedFile{FieldName: key, File: FileUpload{Path: v}})
			case looksLikePath(v) && !isHTTPURL(v):
				return fmt.Errorf("input %s references a file that was not found: %s", key, v)
			default:
				form.Set(key, v)
			}
		case fmt.Stringer:
			form.Set(key, v.String())
		default:
			form.Set(key, fmt.Sprintf("%v", v))
		}
	}

	if len(files) == 0 {
		headers := http.Header{}
		headers.Set("Content-Type", "application/x-www-form-urlencoded")
		data, err := c.doRequest(ctx, http.MethodPost, path, headers, strings.NewReader(form.Encode()), query)
		if err != nil {
			return err
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(data, out)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, values := range form {
		for _, v := range values {
			_ = writer.WriteField(key, v)
		}
	}

	// Track opened readers for cleanup on error
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
			return err
		}
		openedReaders = append(openedReaders, fileReader)

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, f.FieldName, filename))
		h.Set("Content-Type", mimeType)

		part, err := writer.CreatePart(h)
		if err != nil {
			closeAllReaders()
			writer.Close()
			return err
		}
		if _, err := io.Copy(part, fileReader); err != nil {
			closeAllReaders()
			writer.Close()
			return err
		}
		fileReader.Close()
	}
	// Clear the slice since we closed readers individually on success
	openedReaders = nil
	_ = openedReaders // silence unused warning

	if err := writer.Close(); err != nil {
		return err
	}

	headers := http.Header{}
	headers.Set("Content-Type", writer.FormDataContentType())
	data, err := c.doRequest(ctx, http.MethodPost, path, headers, body, query)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
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
