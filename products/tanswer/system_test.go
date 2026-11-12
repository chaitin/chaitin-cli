package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSystemStatusHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "system", "status", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查看系统基础状态", "产品版本", "License", "节点", "自检", "输出"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestSystemStatusCommandCallsBaseInfoLicenseAndStatus(t *testing.T) {
	seen := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		seen[req.Method] = true
		switch req.Method {
		case "OpsService.GetBaseInfo":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"ProductName":"全悉",
					"ProductMode":"分布式版",
					"ProductVersion":"cluster",
					"Version":"6.2.1",
					"InstallTime":"2026-07-01 10:00:00"
				}`),
			})
		case "HeraLicenseService.GetLicense":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"company":"Example Inc",
					"valid_expired":false,
					"valid_expired_soon":true,
					"permanent":false,
					"product_version":"cluster",
					"license_type":"cluster",
					"max_node_count":3,
					"traffic_limit":"1Gbps"
				}`),
			})
		case "OpsService.GetSystemStatusResult":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"status":"running",
					"update_time":"2026-07-17T16:00:00+08:00",
					"online_count":1,
					"total_count":1,
					"nodes":[{"node_id":"node-1","node_name":"管理节点","status":"online","cpu_status":"running","mem_status":"running","disk_status":"running","docker_status":{"alarm_store":"running","upgrader":"running"}}]
				}`),
			})
		default:
			t.Fatalf("unexpected method %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "system", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, method := range []string{"OpsService.GetBaseInfo", "HeraLicenseService.GetLicense", "OpsService.GetSystemStatusResult"} {
		if !seen[method] {
			t.Fatalf("missing RPC call %s; seen=%#v", method, seen)
		}
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查看系统基础状态" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	health := data["health"].(map[string]any)
	if health["status"] != "running" || health["online_count"] != float64(1) {
		t.Fatalf("health = %#v", health)
	}
	version := data["version"].(map[string]any)
	if version["product_version"] != "cluster" || version["version"] != "6.2.1" {
		t.Fatalf("version = %#v", version)
	}
	license := data["license"].(map[string]any)
	if license["valid_expired"] != false || license["valid_expired_soon"] != true {
		t.Fatalf("license = %#v", license)
	}
}
