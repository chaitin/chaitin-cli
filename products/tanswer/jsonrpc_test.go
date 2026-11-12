package tanswer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCallRPCPostsJSONRPC20(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rpc" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.JSONRPC != "2.0" || req.Method != "OpsService.GetBaseInfo" {
			t.Fatalf("bad rpc request: %#v", req)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"product_name":"全悉"}`),
		})
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, APIToken: "token-123", Timeout: time.Second})
	var out map[string]any
	if err := client.CallRPC(context.Background(), "OpsService.GetBaseInfo", map[string]any{}, &out); err != nil {
		t.Fatalf("CallRPC returned error: %v", err)
	}
	if out["product_name"] != "全悉" {
		t.Fatalf("product_name = %#v", out["product_name"])
	}
}
