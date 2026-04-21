package ddr

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chaitin/workspace-cli/config"
)

func TestParseJWTClaims(t *testing.T) {
	claims, err := parseJWTClaims("Bearer e30.eyJVc2VySUQiOiIxIiwiVXNlck5TIjoiZGVmYXVsdCJ9.sig")
	if err != nil {
		t.Fatalf("parseJWTClaims() error = %v", err)
	}
	if claims.UserID != "1" || claims.UserNS != "default" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestBuildServalToken(t *testing.T) {
	got := buildServalToken("69e35caa09f496d33065033a")
	want := "Serval " + base64.StdEncoding.EncodeToString([]byte("serval:69e35caa09f496d33065033a"))
	if got != want {
		t.Fatalf("buildServalToken() = %q, want %q", got, want)
	}
}

func TestNormalizeDDRConfigURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain host", raw: "https://example.com", want: "https://example.com/qzh/api/v1"},
		{name: "already normalized", raw: "https://example.com/qzh/api/v1", want: "https://example.com/qzh/api/v1"},
		{name: "trim trailing slash", raw: "https://example.com/qzh/api/v1/", want: "https://example.com/qzh/api/v1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeDDRConfigURL(tc.raw); got != tc.want {
				t.Fatalf("normalizeDDRConfigURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCreateAndPersistAPIToken(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	runtimeCfg = Config{URL: "https://old.example.com"}

	var seenCreateAuth string
	var seenCreateUser string
	var seenAttrAuth string
	var seenCreatePath string
	var seenAttrPath string

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/qzh/api/auth/v1/access_key/batch":
			seenCreatePath = r.URL.Path
			seenCreateAuth = r.Header.Get("Authorization")
			seenCreateUser = r.Header.Get("User")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"data":[{"access_key":{"access_key":"ak-1","secret_key":"sk-1"}}]}}`))
		case "/qzh/api/auth/v1/ns/attributes":
			seenAttrPath = r.URL.Path
			seenAttrAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"attributes":[{"k":"corp_name","v":"corp-1"}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(&Config{
		URL: server.URL,
	}, nil, false)
	client.httpClient = server.Client()

	result, err := createAndPersistAPIToken(context.Background(), client, configPath, "Bearer e30.eyJVc2VySUQiOiIxIiwiVXNlck5TIjoiZGVmYXVsdCJ9.sig")
	if err != nil {
		t.Fatalf("createAndPersistAPIToken() error = %v", err)
	}

	if seenCreateAuth != "Bearer e30.eyJVc2VySUQiOiIxIiwiVXNlck5TIjoiZGVmYXVsdCJ9.sig" {
		t.Fatalf("create Authorization = %q", seenCreateAuth)
	}
	if seenCreatePath != "/qzh/api/auth/v1/access_key/batch" {
		t.Fatalf("create path = %q", seenCreatePath)
	}
	if !strings.Contains(seenCreateUser, `"user_id":"1"`) || !strings.Contains(seenCreateUser, `"ns":"default"`) {
		t.Fatalf("User header = %q", seenCreateUser)
	}

	wantToken := base64.StdEncoding.EncodeToString([]byte("serval:ak-1"))
	if seenAttrAuth != wantToken {
		t.Fatalf("ns attributes Authorization = %q, want %q", seenAttrAuth, wantToken)
	}
	if seenAttrPath != "/qzh/api/auth/v1/ns/attributes" {
		t.Fatalf("attributes path = %q", seenAttrPath)
	}
	if result.Token != wantToken || result.CompanyID != "corp-1" {
		t.Fatalf("unexpected result: %+v", result)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var saved Config
	node := cfg["ddr"]
	if err := node.Decode(&saved); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if saved.URL != server.URL+"/qzh/api/v1" || saved.APIKey != wantToken || saved.CompanyID != "corp-1" {
		t.Fatalf("unexpected saved config: %+v", saved)
	}
}
