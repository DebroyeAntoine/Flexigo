package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func expandEnvVars(text string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match
	})
}

func (h *HTTPClient) ExecuteRequest(method, url string, headers map[string]string, body string) error {
	url = expandEnvVars(url)
	body = expandEnvVars(body)

	var req *http.Request
	var err error

	if body != "" {
		req, err = http.NewRequest(method, url, bytes.NewBufferString(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		expandedValue := expandEnvVars(value)
		req.Header.Set(key, expandedValue)
	}

	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error %d: %s\nResponse: %s",
			resp.StatusCode, resp.Status, string(responseBody))
	}

	fmt.Printf("HTTP %s %s -> %d %s\n", method, url, resp.StatusCode, resp.Status)
	return nil
}
