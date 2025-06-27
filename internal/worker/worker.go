package worker

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"stampede-shooter/internal/config"
	"stampede-shooter/internal/metrics"
	"stampede-shooter/internal/script"
	"stampede-shooter/internal/util"
)

// Worker represents a single virtual user
type Worker struct {
	id             int
	client         *http.Client
	rateLimiter    *util.RateLimiter
	script         *script.Script
	collector      *metrics.Collector
	loginHeader    string
	sessionHeaders map[string]string        // Persistent headers across requests
	csrfToken      string                   // Current CSRF token for Rails apps
	credentials    *util.CredentialsManager // Credentials manager for authentication
}

// New creates a new worker
func New(id int, cfg config.Config, script *script.Script, collector *metrics.Collector, credentials *util.CredentialsManager) *Worker {
	// Configure HTTP client with cookie jar for session persistence
	jar, _ := cookiejar.New(nil)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
	}

	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		Jar:       jar, // Enable cookie persistence
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects (default behavior)
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	return &Worker{
		id:             id,
		client:         client,
		rateLimiter:    util.NewRateLimiter(cfg.RPS),
		script:         script,
		collector:      collector,
		loginHeader:    cfg.LoginHeader,
		sessionHeaders: make(map[string]string),
		credentials:    credentials,
	}
}

// Run executes the worker's test script
func (w *Worker) Run(ctx context.Context, loginURL string) error {
	// Optional login step
	if loginURL != "" {
		if err := w.login(ctx, loginURL); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	// Execute script actions in a loop until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := w.executeScript(ctx); err != nil {
				// Log error but continue running
				continue
			}
		}
	}
}

// login performs the optional login request
func (w *Worker) login(ctx context.Context, loginURL string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, nil)
	if err != nil {
		return err
	}

	// Add login header if provided
	if w.loginHeader != "" {
		parts := strings.SplitN(w.loginHeader, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	// Store any session headers from login response
	w.extractSessionHeaders(resp)

	return nil
}

// executeScript runs through all actions in the script once
func (w *Worker) executeScript(ctx context.Context) error {
	for _, action := range w.script.Actions {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Rate limit requests
			w.rateLimiter.Wait()

			// Execute action
			w.executeAction(ctx, action)

			// Apply delay after action (except for the last action)
			if delay := action.GetDelay(); delay > 0 {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(delay):
					// Delay completed, continue to next action
				}
			}
		}
	}
	return nil
}

// executeAction performs a single HTTP action
func (w *Worker) executeAction(ctx context.Context, action script.Action) {
	// Expand templates with user-specific data
	expandedAction := action.ExpandTemplates(w.id)

	// Replace credential placeholders if credentials manager is available
	if w.credentials != nil {
		creds := w.credentials.GetCredentialsForUser(w.id)
		expandedAction.Body = w.replaceCredentialPlaceholders(expandedAction.Body, creds)
		expandedAction.JSONBody = w.replaceCredentialPlaceholders(expandedAction.JSONBody, creds)
	}

	startTime := time.Now()

	// Create request
	var body io.Reader
	if expandedAction.JSONBody != "" {
		body = bytes.NewBufferString(expandedAction.JSONBody)
	} else if expandedAction.Body != "" {
		// Replace CSRF token placeholder in body if present
		bodyContent := expandedAction.Body
		if w.csrfToken != "" {
			// URL-encode the CSRF token for form data
			encodedToken := url.QueryEscape(w.csrfToken)
			bodyContent = strings.ReplaceAll(bodyContent, "CSRF_TOKEN_PLACEHOLDER", encodedToken)
		}
		body = bytes.NewBufferString(bodyContent)
	}

	req, err := http.NewRequestWithContext(ctx, expandedAction.Method, expandedAction.URL, body)
	if err != nil {
		w.recordMetric(expandedAction, startTime, time.Now(), 0, 0, err.Error())
		return
	}

	// Set content type for JSON requests
	if expandedAction.JSONBody != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers from script
	for key, value := range expandedAction.Headers {
		// Skip Accept-Encoding to let Go handle decompression automatically
		if key == "Accept-Encoding" {
			continue
		}
		req.Header.Set(key, value)
	}

	// Add persistent session headers
	for key, value := range w.sessionHeaders {
		req.Header.Set(key, value)
	}

	// Add CSRF token header if we have one
	if w.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", w.csrfToken)
	}

	// Add login header if provided
	if w.loginHeader != "" {
		parts := strings.SplitN(w.loginHeader, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Execute request
	resp, err := w.client.Do(req)
	endTime := time.Now()

	if err != nil {
		w.recordMetric(expandedAction, startTime, endTime, 0, 0, err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body (Go automatically handles decompression when Accept-Encoding is not set)
	bodyBytes, _ := io.ReadAll(resp.Body)
	bytesRead := int64(len(bodyBytes))

	// Extract CSRF token from HTML response if this is a login page
	if strings.Contains(expandedAction.URL, "sign_in") || strings.Contains(expandedAction.URL, "login") {
		w.extractCSRFTokenFromHTML(string(bodyBytes))
	}

	// Extract and store any new session headers
	w.extractSessionHeaders(resp)

	// Check expected status
	errorMsg := ""
	if expandedAction.ExpectStatus > 0 && resp.StatusCode != expandedAction.ExpectStatus {
		errorMsg = fmt.Sprintf("expected status %d, got %d", expandedAction.ExpectStatus, resp.StatusCode)
	}

	w.recordMetric(expandedAction, startTime, endTime, resp.StatusCode, bytesRead, errorMsg)
}

// replaceCredentialPlaceholders replaces credential placeholders in request bodies
func (w *Worker) replaceCredentialPlaceholders(content string, creds util.Credentials) string {
	if content == "" {
		return content
	}

	// Replace username and password placeholders
	content = strings.ReplaceAll(content, "{{username}}", creds.Username)
	content = strings.ReplaceAll(content, "{{password}}", creds.Password)

	// Also support email format for Rails apps
	content = strings.ReplaceAll(content, "{{email}}", creds.Username)

	return content
}

// extractCSRFTokenFromHTML extracts CSRF token from HTML response
func (w *Worker) extractCSRFTokenFromHTML(htmlContent string) {
	// Method 1: Extract from meta tag
	metaPattern := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)"`)
	if matches := metaPattern.FindStringSubmatch(htmlContent); len(matches) > 1 {
		w.csrfToken = matches[1]
		return
	}

	// Method 2: Extract from form input
	formPattern := regexp.MustCompile(`<input[^>]*name="authenticity_token"[^>]*value="([^"]+)"`)
	if matches := formPattern.FindStringSubmatch(htmlContent); len(matches) > 1 {
		w.csrfToken = matches[1]
		return
	}

	// Method 3: Extract from any input with authenticity_token
	authPattern := regexp.MustCompile(`authenticity_token"[^>]*value="([^"]+)"`)
	if matches := authPattern.FindStringSubmatch(htmlContent); len(matches) > 1 {
		w.csrfToken = matches[1]
		return
	}
}

// extractSessionHeaders extracts important headers from response for future requests
func (w *Worker) extractSessionHeaders(resp *http.Response) {
	// Extract CSRF token from response headers
	if csrfToken := resp.Header.Get("X-CSRF-Token"); csrfToken != "" {
		w.sessionHeaders["X-CSRF-Token"] = csrfToken
		w.csrfToken = csrfToken // Also update the current CSRF token
	}

	// Extract other session-related headers as needed
	if authHeader := resp.Header.Get("Authorization"); authHeader != "" {
		w.sessionHeaders["Authorization"] = authHeader
	}

	// Note: Cookies are automatically handled by the cookie jar
}

// recordMetric sends a metric to the collector
func (w *Worker) recordMetric(action script.Action, start, end time.Time, statusCode int, bytesRead int64, errorMsg string) {
	metric := metrics.RequestMetric{
		Name:       action.Name,
		Method:     action.Method,
		URL:        action.URL,
		StartTime:  start,
		EndTime:    end,
		StatusCode: statusCode,
		BytesRead:  bytesRead,
		Error:      errorMsg,
	}

	w.collector.Record(metric)
}
