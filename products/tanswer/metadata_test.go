package tanswer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetadataProtocolHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "metadata", "protocol", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"按协议检索元数据", "HTTP", "DNS", "--protocol", "--page-size", "输出"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildMetadataListRequestMapsFilters(t *testing.T) {
	rng := TimeRange{Start: 1000, End: 2000}
	req := buildMetadataListRequest(rng, metadataListOptions{
		page:          2,
		pageSize:      20,
		protocol:      "http",
		srcIP:         "198.51.100.10",
		destIP:        "192.0.2.10",
		srcPort:       "12345",
		destPort:      "443",
		advancedQuery: "http_method = 'GET'",
	})

	if req["time_range_start"] != int64(1000) || req["time_range_end"] != int64(2000) {
		t.Fatalf("time range mismatch: %#v", req)
	}
	if req["offset"] != int64(20) || req["count"] != int64(20) {
		t.Fatalf("pagination mismatch: %#v", req)
	}
	if req["event_type"] != "http" || req["advanced_query"] != "http_method = 'GET'" {
		t.Fatalf("event/advanced mismatch: %#v", req)
	}
	if got := req["src_ip"].([]map[string]any)[0]["target"]; got != "198.51.100.10" {
		t.Fatalf("src_ip target = %#v", got)
	}
	if got := req["dest_port"].([]map[string]any)[0]["target"]; got != 443 {
		t.Fatalf("dest_port target = %#v", got)
	}
}

func TestMetadataProtocolCommandCallsHTTPLogSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "LogSearchService.SearchOrigDataHTTPLog" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["event_type"] != "http" || params["count"] != float64(5) || params["offset"] != float64(0) {
			t.Fatalf("params = %#v", params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"logs":[{"id":"m1","timestamp":1784282400000,"event_type":"http","src_ip":"198.51.100.10","src_port":12345,"dest_ip":"192.0.2.10","dest_port":80,"proto":"tcp","app_proto":"http","http":{"hostname":"example.com","url":"/login","req_line":"GET /login HTTP/1.1","resp_line":"HTTP/1.1 200 OK","user_agent":"curl"}}],
				"total":1
			}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "metadata", "protocol", "--protocol", "http", "--time", "24h", "--page-size", "5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "按协议检索元数据" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	items := data["metadata"].([]any)
	item := items[0].(map[string]any)
	if item["id"] != "m1" || item["hostname"] != "example.com" || item["url"] != "/login" {
		t.Fatalf("summary mismatch: %#v", item)
	}
}

func TestMetadataSearchCommandPassesAdvancedQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "LogSearchService.SearchOrigDataDNSLog" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["advanced_query"] != "dns_rrname = 'example.com'" {
			t.Fatalf("advanced_query = %#v", params["advanced_query"])
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"logs":[{"id":"d1","timestamp":1784282400000,"event_type":"dns","src_ip":"198.51.100.10","dest_ip":"203.0.113.53","app_proto":"dns","dns":{"rrname":"example.com","rrtype":"A","rcode":"NOERROR"}}],"total":1}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "metadata", "search", "--protocol", "dns", "--advanced-query", "dns_rrname = 'example.com'"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "按高级条件检索元数据" {
		t.Fatalf("Task = %q", env.Task)
	}
}

func TestMetadataSearchRequiresAdvancedQuery(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "metadata", "search", "--protocol", "dns"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "MISSING_ADVANCED_QUERY" {
		t.Fatalf("error code = %q, output:\n%s", env.Error.Code, out.String())
	}
}

func TestMetadataDetailCommandCallsDetailRPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "LogSearchService.GetOrigDataLogDetail" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["id"] != "m1" || params["timestamp"] != float64(1784282400000) || params["event_type_group"] != "http" {
			t.Fatalf("params = %#v", params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"log":{"id":"m1","event_type":"http","timestamp":1784282400000,"src_ip":"198.51.100.10","http":{"url":"/login"}}}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "metadata", "detail", "--id", "m1", "--timestamp", "1784282400000", "--protocol", "http"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查看元数据详情" {
		t.Fatalf("Task = %q", env.Task)
	}
}

func TestMetadataConfigCommandCallsProtocolList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "LogSearchService.GetOrigDataLogProtocolList" {
			t.Fatalf("method = %q", req.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"configurations":[{"event_type":"http","selected":true,"display_field":"default"},{"event_type":"dns","selected":true,"display_field":"default"}]}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "metadata", "config"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询元数据数据配置" {
		t.Fatalf("Task = %q", env.Task)
	}
}

func TestMetadataConfigUpdateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "metadata", "config-update", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"调整元数据数据配置", "--node-id", "--enable", "--disable", "--confirm", "CONFIRM_METADATA_CONFIG_UPDATE", "输出预览"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildMetadataConfigUpdateRequestMapsEnabledProtocols(t *testing.T) {
	req, err := buildMetadataConfigUpdateRequest([]map[string]any{
		{"event_type": "http", "display_field": "HTTP", "selected": false},
		{"event_type": "dns", "display_field": "DNS", "selected": true},
		{"event_type": "tcp", "display_field": "TCP", "selected": true},
	}, metadataConfigWriteOptions{
		nodeID:  "node-1",
		enable:  "http",
		disable: "tcp",
	})
	if err != nil {
		t.Fatalf("buildMetadataConfigUpdateRequest returned error: %v", err)
	}
	if req["node_id"] != "node-1" {
		t.Fatalf("node_id = %#v", req["node_id"])
	}
	configs := req["configurations"].([]map[string]any)
	if len(configs) != 3 {
		t.Fatalf("configurations = %#v", configs)
	}
	got := map[string]bool{}
	for _, item := range configs {
		got[item["event_type"].(string)] = item["selected"].(bool)
	}
	if !got["http"] || !got["dns"] || got["tcp"] {
		t.Fatalf("selected map = %#v", got)
	}
}

func TestMetadataConfigUpdatePreviewReadsCurrentConfigOnly(t *testing.T) {
	server := newMetadataRPCSequenceServer(t, []metadataExpectedRPC{{
		method: "LogSearchService.GetOrigDataLogProtocolList",
		check: func(t *testing.T, params map[string]any) {
			if params["node_id"] != "node-1" {
				t.Fatalf("node_id = %#v", params["node_id"])
			}
		},
		result: `{"configurations":[{"event_type":"http","selected":false,"display_field":"HTTP"},{"event_type":"dns","selected":true,"display_field":"DNS"}]}`,
	}})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "metadata", "config-update", "--node-id", "node-1", "--enable", "http", "--preview"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 1 {
		t.Fatalf("request count = %d", server.requestCount)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "调整元数据数据配置预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmation_token"] != "CONFIRM_METADATA_CONFIG_UPDATE" || data["confirmed"] != false {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestMetadataConfigUpdateConfirmedCallsConfigure(t *testing.T) {
	server := newMetadataRPCSequenceServer(t, []metadataExpectedRPC{
		{
			method: "LogSearchService.GetOrigDataLogProtocolList",
			result: `{"configurations":[{"event_type":"http","selected":false,"display_field":"HTTP"},{"event_type":"dns","selected":true,"display_field":"DNS"}]}`,
		},
		{
			method: "LogSearchService.ConfigureOrigDataLogProtocol",
			check: func(t *testing.T, params map[string]any) {
				if params["node_id"] != "node-1" {
					t.Fatalf("node_id = %#v", params["node_id"])
				}
				configs := params["configurations"].([]any)
				if len(configs) != 2 {
					t.Fatalf("configurations = %#v", configs)
				}
				selected := map[string]bool{}
				for _, raw := range configs {
					item := raw.(map[string]any)
					selected[item["event_type"].(string)] = item["selected"].(bool)
				}
				if !selected["http"] || selected["dns"] {
					t.Fatalf("selected = %#v", selected)
				}
			},
			result: `{}`,
		},
	})
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "metadata", "config-update", "--node-id", "node-1", "--enable", "http", "--disable", "dns", "--confirm", "CONFIRM_METADATA_CONFIG_UPDATE"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if server.requestCount != 2 {
		t.Fatalf("request count = %d", server.requestCount)
	}
}

func TestMetadataNearAlarmCombinesAlarmDetailAndMetadataSearch(t *testing.T) {
	seen := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		seen[req.Method] = true
		switch req.Method {
		case "AlarmService.GetAlarm":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":{"doc_id":"alarm-1","timestamp":1784282400000,"src_ip":"198.51.100.10","src_port":12345,"dest_ip":"192.0.2.10","dest_port":80,"app_proto":"http"}}`),
			})
		case "LogSearchService.SearchOrigDataHTTPLog":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"logs":[{"id":"m1","timestamp":1784282400001,"event_type":"http","src_ip":"198.51.100.10","src_port":12345,"dest_ip":"192.0.2.10","dest_port":80,"app_proto":"http","http":{"url":"/attack"}}],"total":1}`),
			})
		default:
			t.Fatalf("unexpected method %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "--url", server.URL, "--api-key", "token-123", "metadata", "near-alarm", "--id", "alarm-1", "--window", "10m", "--page-size", "5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !seen["AlarmService.GetAlarm"] || !seen["LogSearchService.SearchOrigDataHTTPLog"] {
		t.Fatalf("missing expected calls: %#v", seen)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询告警附近元数据" {
		t.Fatalf("Task = %q", env.Task)
	}
}

type metadataExpectedRPC struct {
	method string
	check  func(t *testing.T, params map[string]any)
	result string
}

type metadataRPCSequenceServer struct {
	*httptest.Server
	requestCount int
}

func newMetadataRPCSequenceServer(t *testing.T, expected []metadataExpectedRPC) *metadataRPCSequenceServer {
	t.Helper()
	state := &metadataRPCSequenceServer{}
	state.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if state.requestCount >= len(expected) {
			t.Fatalf("unexpected extra RPC request")
		}
		want := expected[state.requestCount]
		state.requestCount++
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != want.method {
			t.Fatalf("method = %q, want %q", req.Method, want.method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if want.check != nil {
			want.check(t, params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(want.result),
		})
	}))
	return state
}
