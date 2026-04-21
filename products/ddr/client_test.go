package ddr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientInjectsAuthHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "test-token" {
			t.Fatalf("Authorization = %q, want %q", got, "test-token")
		}
		if got := r.Header.Get("X-CS-Header-Company"); got != "company-1" {
			t.Fatalf("X-CS-Header-Company = %q, want %q", got, "company-1")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"items":[]}}`))
	}))
	defer server.Close()

	client := NewClient(&Config{
		URL:       server.URL,
		APIKey:    "test-token",
		CompanyID: "company-1",
	}, nil, false)

	var result map[string]interface{}
	if err := client.Do(context.Background(), http.MethodGet, "/health", nil, nil, nil, &result); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}
