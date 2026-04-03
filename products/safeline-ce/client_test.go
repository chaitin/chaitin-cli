package safelinece

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	cfg := &Config{
		Endpoint: "https://example.com:9443",
		Token:    "test-token",
	}
	client := NewClient(cfg)

	if client.baseURL != "https://example.com:9443" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://example.com:9443")
	}
	if client.config.Token != "test-token" {
		t.Errorf("Token = %q, want %q", client.config.Token, "test-token")
	}
}

func TestClientDo(t *testing.T) {
	t.Run("successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-SLCE-API-TOKEN") != "test-token" {
				t.Error("missing X-SLCE-API-TOKEN header")
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": {"id": 1, "name": "test"}}`))
		}))
		defer server.Close()

		cfg := &Config{Endpoint: server.URL, Token: "test-token"}
		client := NewClient(cfg)

		var result map[string]interface{}
		err := client.Get(context.Background(), "/test", nil, &result)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("response data is not a map")
		}
		if data["id"].(float64) != 1 {
			t.Errorf("data.id = %v, want 1", data["id"])
		}
	})

	t.Run("API error 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors": [{"code": "NOT_FOUND", "message": "resource not found"}]}`))
		}))
		defer server.Close()

		cfg := &Config{Endpoint: server.URL, Token: "test-token"}
		client := NewClient(cfg)

		err := client.Get(context.Background(), "/test", nil, nil)
		if err == nil {
			t.Fatal("Get() expected error for 404")
		}

		cliErr, ok := err.(*CLIError)
		if !ok {
			t.Fatalf("error type = %T, want *CLIError", err)
		}
		if cliErr.Code != ExitAPIError {
			t.Errorf("error code = %d, want %d", cliErr.Code, ExitAPIError)
		}
		if cliErr.StatusCode != 404 {
			t.Errorf("status code = %d, want 404", cliErr.StatusCode)
		}
	})

	t.Run("network error", func(t *testing.T) {
		cfg := &Config{
			Endpoint: "http://nonexistent-host-12345:9999",
			Token:    "test-token",
		}
		client := NewClient(cfg)
		client.httpClient.Timeout = 1 * time.Second

		err := client.Get(context.Background(), "/test", nil, nil)
		if err == nil {
			t.Fatal("Get() expected error for network failure")
		}

		cliErr, ok := err.(*CLIError)
		if !ok {
			t.Fatalf("error type = %T, want *CLIError", err)
		}
		if cliErr.Code != ExitNetworkError {
			t.Errorf("error code = %d, want %d", cliErr.Code, ExitNetworkError)
		}
	})

	t.Run("POST request with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Method = %s, want POST", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
			}

			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			if req["name"] != "test" {
				t.Errorf("request body name = %v, want test", req["name"])
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": {"id": 1}}`))
		}))
		defer server.Close()

		cfg := &Config{Endpoint: server.URL, Token: "test-token"}
		client := NewClient(cfg)

		body := map[string]interface{}{"name": "test"}
		var result map[string]interface{}
		err := client.Post(context.Background(), "/test", body, &result)
		if err != nil {
			t.Fatalf("Post() error = %v", err)
		}
	})
}

func TestClient_TokenInjection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-SLCE-API-TOKEN")
		if token != "secret-token" {
			t.Errorf("X-SLCE-API-TOKEN = %q, want %q", token, "secret-token")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	cfg := &Config{Endpoint: server.URL, Token: "secret-token"}
	client := NewClient(cfg)

	tests := []struct {
		name string
		call func() error
	}{
		{"GET", func() error { return client.Get(context.Background(), "/test", nil, nil) }},
		{"POST", func() error { return client.Post(context.Background(), "/test", map[string]string{}, nil) }},
		{"PUT", func() error { return client.Put(context.Background(), "/test", map[string]string{}, nil) }},
		{"DELETE", func() error { return client.Delete(context.Background(), "/test", nil) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err != nil {
				t.Errorf("%s() error = %v", tc.name, err)
			}
		})
	}
}

func TestClient_BuildURL(t *testing.T) {
	cfg := &Config{Endpoint: "https://example.com", Token: "token"}
	client := NewClient(cfg)

	tests := []struct {
		path string
		want string
	}{
		{"/api/test", "https://example.com/api/test"},
		{"api/test", "https://example.com/api/test"},
		{"/api/test/", "https://example.com/api/test/"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := client.buildURL(tt.path)
			if got != tt.want {
				t.Errorf("buildURL(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestClient_QueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("page") != "1" {
			t.Errorf("page = %q, want 1", query.Get("page"))
		}
		if query.Get("size") != "20" {
			t.Errorf("size = %q, want 20", query.Get("size"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	cfg := &Config{Endpoint: server.URL, Token: "token"}
	client := NewClient(cfg)

	query := url.Values{}
	query.Set("page", "1")
	query.Set("size", "20")

	err := client.Get(context.Background(), "/test", query, nil)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestClient_HandleError(t *testing.T) {
	cfg := &Config{Endpoint: "https://example.com", Token: "token"}
	client := NewClient(cfg)

	t.Run("API error with message", func(t *testing.T) {
		body := []byte(`{"errors": [{"code": "INVALID", "message": "invalid input"}]}`)
		err := client.handleError(400, body)

		cliErr, ok := err.(*CLIError)
		if !ok {
			t.Fatalf("error type = %T, want *CLIError", err)
		}
		if cliErr.Code != ExitAPIError {
			t.Errorf("Code = %d, want %d", cliErr.Code, ExitAPIError)
		}
		if !strings.Contains(cliErr.Message, "invalid input") {
			t.Errorf("Message = %q, should contain %q", cliErr.Message, "invalid input")
		}
	})

	t.Run("API error without message", func(t *testing.T) {
		body := []byte(`{}`)
		err := client.handleError(500, body)

		cliErr, ok := err.(*CLIError)
		if !ok {
			t.Fatalf("error type = %T, want *CLIError", err)
		}
		if cliErr.Code != ExitAPIError {
			t.Errorf("Code = %d, want %d", cliErr.Code, ExitAPIError)
		}
		if !strings.Contains(cliErr.Message, "500") {
			t.Errorf("Message = %q, should contain status code", cliErr.Message)
		}
	})
}
