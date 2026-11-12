package cloudatlas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientSuggestsOpenAPIBaseURLForConsoleHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!doctype html><html></html>"))
	}))
	defer server.Close()

	client := NewClient(Config{URL: server.URL}, false, false)
	_, err := client.Do(context.Background(), http.MethodGet, "/v1/risk/high-risk", nil, nil)
	if err == nil {
		t.Fatal("Do() error = nil, want HTML response error")
	}
	want := "try --url " + server.URL + "/openapi"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("Do() error = %q, want it to contain %q", err, want)
	}
}

func TestClientSuggestsOpenAPIBaseURLForBrowserSessionAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":401,"message":"用户未登录","data":null}`))
	}))
	defer server.Close()

	client := NewClient(Config{URL: server.URL + "/api", Token: "token"}, false, false)
	_, err := client.Do(context.Background(), http.MethodGet, "/v1/risk/high-risk", nil, nil)
	if err == nil {
		t.Fatal("Do() error = nil, want authentication error")
	}
	want := "try --url " + server.URL + "/openapi"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("Do() error = %q, want it to contain %q", err, want)
	}
}

func TestClientDoesNotGuessOpenAPIBaseURLForCustomPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("<!doctype html><html></html>"))
	}))
	defer server.Close()

	client := NewClient(Config{URL: server.URL + "/custom"}, false, false)
	_, err := client.Do(context.Background(), http.MethodGet, "/v1/risk/high-risk", nil, nil)
	if err == nil {
		t.Fatal("Do() error = nil, want HTML response error")
	}
	if strings.Contains(err.Error(), "try --url") {
		t.Fatalf("Do() error = %q, should not guess a base URL for a custom path", err)
	}
	if want := "status 503"; !strings.Contains(err.Error(), want) {
		t.Fatalf("Do() error = %q, want it to retain %q", err, want)
	}
}
