package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHTTPClient_ExecuteRequest_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewHTTPClient()
	err := client.ExecuteRequest("GET", server.URL, nil, "")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
}

func TestHTTPClient_ExecuteRequest_POST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient()
	body := `{"entity_id": "light.salon"}`
	err := client.ExecuteRequest("POST", server.URL, nil, body)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
}

func TestHTTPClient_ExecuteRequest_WithHeaders(t *testing.T) {
	expectedToken := "secret-token-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+expectedToken {
			t.Errorf("Expected Authorization header 'Bearer %s', got '%s'", expectedToken, auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient()
	headers := map[string]string{
		"Authorization": "Bearer " + expectedToken,
	}
	err := client.ExecuteRequest("GET", server.URL, headers, "")
	if err != nil {
		t.Fatalf("Request with headers failed: %v", err)
	}
}

func TestHTTPClient_ExecuteRequest_EnvVarExpansion(t *testing.T) {
	os.Setenv("TEST_TOKEN", "my-secret-token")
	defer os.Unsetenv("TEST_TOKEN")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-secret-token" {
			t.Errorf("Expected expanded token, got '%s'", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient()
	headers := map[string]string{
		"Authorization": "Bearer ${TEST_TOKEN}",
	}
	err := client.ExecuteRequest("GET", server.URL, headers, "")
	if err != nil {
		t.Fatalf("Request with env var failed: %v", err)
	}
}

func TestHTTPClient_ExecuteRequest_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client := NewHTTPClient()
	err := client.ExecuteRequest("GET", server.URL, nil, "")
	if err == nil {
		t.Fatal("Expected error for 404 status, got nil")
	}
}

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "Single variable",
			input:    "Bearer ${TOKEN}",
			envVars:  map[string]string{"TOKEN": "abc123"},
			expected: "Bearer abc123",
		},
		{
			name:     "Multiple variables",
			input:    "${PROTOCOL}://${HOST}:${PORT}",
			envVars:  map[string]string{"PROTOCOL": "https", "HOST": "example.com", "PORT": "8080"},
			expected: "https://example.com:8080",
		},
		{
			name:     "No variables",
			input:    "static text",
			envVars:  map[string]string{},
			expected: "static text",
		},
		{
			name:     "Undefined variable",
			input:    "Bearer ${UNDEFINED}",
			envVars:  map[string]string{},
			expected: "Bearer ${UNDEFINED}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
