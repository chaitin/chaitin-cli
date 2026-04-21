package ddr

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientInjectsAuthHeaders(t *testing.T) {
	client := NewClient(&Config{
		URL:       "https://example.test",
		APIKey:    "test-token",
		CompanyID: "company-1",
	}, nil, false)
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.Header.Get("Authorization"); got != "Serval test-token" {
				t.Fatalf("Authorization = %q, want %q", got, "Serval test-token")
			}
			if got := req.Header.Get("X-CS-Header-Company"); got != "company-1" {
				t.Fatalf("X-CS-Header-Company = %q, want %q", got, "company-1")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"msg":"ok","data":{"items":[]}}`)),
				Request:    req,
			}, nil
		}),
	}

	var result map[string]interface{}
	if err := client.Do(context.Background(), http.MethodGet, "/health", nil, nil, nil, &result); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}
