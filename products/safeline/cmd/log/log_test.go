package log

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cmdpkg "github.com/chaitin/chaitin-cli/products/safeline/cmd"
	"github.com/chaitin/chaitin-cli/products/safeline/pkg/client"
)

func TestDetectGetDoesNotRequireTimestamp(t *testing.T) {
	testDetectGetOmitsEmptyTimestamp(t, []string{"--event-id", "event-1"})
}

func TestDetectGetOmitsExplicitEmptyTimestamp(t *testing.T) {
	testDetectGetOmitsEmptyTimestamp(t, []string{"--event-id", "event-1", "--timestamp", ""})
}

func TestDetectListAddsSupportedFilters(t *testing.T) {
	wantQuery := map[string]string{
		"scope":                           "log:detect_log:optim_limit",
		"exclude_body":                    "true",
		"current_page":                    "0",
		"target_page":                     "1",
		"count":                           "20",
		"src_ip__exact":                   "1.1.1.1",
		"src_ip__net_contained_or_equal":  "10.0.0.0/24",
		"host__contains":                  "example.com",
		"url_path__contains":              "/login",
		"method__in":                      "GET",
		"status_code__exact":              "403",
		"event_id__exact":                 "event-1",
		"risk_level__in":                  "2",
		"dest_ip__net_contained_or_equal": "192.168.0.0/16",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/FilterV2API" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		for k, want := range wantQuery {
			if got := q.Get(k); got != want {
				t.Fatalf("%s = %q, want %q", k, got, want)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(client.Envelope{
			Data: json.RawMessage(`{"items":[],"total":0}`),
		})
	}))
	defer srv.Close()

	cmdpkg.SetFlags(srv.URL, "token", "json", false, false, "", "", "", "")

	c := newDetectListCmd()
	out := &strings.Builder{}
	c.SetOut(out)
	c.SetErr(&strings.Builder{})
	c.SetArgs([]string{
		"--src-ip", "1.1.1.1",
		"--src-cidr", "10.0.0.0/24",
		"--host", "example.com",
		"--url-path", "/login",
		"--method", "GET",
		"--status-code", "403",
		"--event-id", "event-1",
		"--filter", "risk_level__in=2",
		"--filter", "dest_ip__net_contained_or_equal=192.168.0.0/16",
	})

	if err := c.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func testDetectGetOmitsEmptyTimestamp(t *testing.T, args []string) {
	t.Helper()

	var sawTimestamp bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/FilterV2API" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("scope"); got != "log:detect_log:detail" {
			t.Fatalf("scope = %q, want log:detect_log:detail", got)
		}
		if got := q.Get("event_id__exact"); got != "event-1" {
			t.Fatalf("event_id__exact = %q, want event-1", got)
		}
		_, sawTimestamp = q["timestamp__exact"]

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(client.Envelope{
			Data: json.RawMessage(`[{"event_id":"event-1","timestamp":"123","src_ip":"1.1.1.1"}]`),
		})
	}))
	defer srv.Close()

	cmdpkg.SetFlags(srv.URL, "token", "json", false, false, "", "", "", "")

	c := newDetectGetCmd()
	out := &strings.Builder{}
	c.SetOut(out)
	c.SetErr(&strings.Builder{})
	c.SetArgs(args)

	if err := c.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if sawTimestamp {
		t.Fatal("timestamp__exact was sent without --timestamp")
	}
}
