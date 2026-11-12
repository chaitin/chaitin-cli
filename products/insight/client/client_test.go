package client

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chaitin/chaitin-cli/products/insight/models"
)

func TestRequestRejectsNonZeroBusinessCodeOnHTTP200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":100,"msg":"任务ID不能为空","data":null}`))
	}))
	t.Cleanup(server.Close)

	var out, errOut bytes.Buffer
	c := NewClient(models.Config{URL: server.URL}, true, false, false, &out, &errOut)

	_, err := c.Request(http.MethodGet, "/exposure/api/result", nil)
	if err == nil {
		t.Fatal("Request() error = nil, want business code error")
	}
	if !strings.Contains(err.Error(), "任务ID不能为空") {
		t.Fatalf("Request() error = %q, want message", err.Error())
	}
	if !strings.Contains(err.Error(), "Code 100") {
		t.Fatalf("Request() error = %q, want code", err.Error())
	}
}

func TestRequestAllowsZeroBusinessCodeOnHTTP200(t *testing.T) {
	const body = `{"code":0,"msg":"","data":{"ok":true}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	var out, errOut bytes.Buffer
	c := NewClient(models.Config{URL: server.URL}, true, false, false, &out, &errOut)

	got, err := c.Request(http.MethodGet, "/exposure/api/result?id=1", nil)
	if err != nil {
		t.Fatalf("Request() error = %v, want nil", err)
	}
	if string(got) != body {
		t.Fatalf("Request() body = %q, want %q", string(got), body)
	}
}

func TestRequestAllowsResponsesWithoutBusinessCode(t *testing.T) {
	const body = `{"status":"ok"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	var out, errOut bytes.Buffer
	c := NewClient(models.Config{URL: server.URL}, true, false, false, &out, &errOut)

	got, err := c.Request(http.MethodGet, "/static/version", nil)
	if err != nil {
		t.Fatalf("Request() error = %v, want nil", err)
	}
	if string(got) != body {
		t.Fatalf("Request() body = %q, want %q", string(got), body)
	}
}
