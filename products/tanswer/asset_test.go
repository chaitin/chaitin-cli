package tanswer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetListHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "list", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查询资产列表", "资产名称", "--page-size", "--ip", "--asset-type", "--importance"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetGroupTreeHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "group-tree", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"查询资产组树", "资产组 ID", "--id", "--depth", "--with-asset"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetDownloadTemplateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "download-template", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"下载资产导入模板", "--output", "--with-example", "文件"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetExportHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "export", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"导出资产", "--id-list", "--output", "文件"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetCreateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "create", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"新增资产", "--name", "--ip", "--preview", "--confirm"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetUpdateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "update", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"编辑资产", "--id", "--name", "--ip", "--preview", "--confirm"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetDeleteHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "delete", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"删除资产", "--id-list", "--preview", "--confirm", "CONFIRM_ASSET_DELETE"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetBatchMaintainHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "batch-maintain", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"批量维护资产", "--id-list", "--contact", "--location", "--remark", "--group-id", "CONFIRM_ASSET_BATCH_MAINTAIN"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetBatchTagHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "batch-tag", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"批量维护资产标签", "--id-list", "--tag-id", "--preview", "--confirm", "CONFIRM_ASSET_BATCH_TAG"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetGroupCreateHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "group-create", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"创建资产组", "--name", "--parent-id", "--preview", "--confirm", "CONFIRM_ASSET_GROUP_CREATE"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetGroupRenameHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "group-rename", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"重命名资产组", "--id", "--name", "--preview", "--confirm", "CONFIRM_ASSET_GROUP_RENAME"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetGroupDeleteHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "group-delete", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"删除资产组", "--id-list", "--preview", "--confirm", "CONFIRM_ASSET_GROUP_DELETE"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetTreeMoveHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "tree-move", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"调整资产树层级", "--id", "--type", "--prev-id", "--prev-type", "--top-layer", "CONFIRM_ASSET_TREE_MOVE"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestAssetImportHelpIsAIReadable(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{"tanswer", "asset", "import", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{"导入资产", "--file", "--preview", "--confirm", "CONFIRM_ASSET_IMPORT"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q:\n%s", want, text)
		}
	}
}

func TestBuildAssetListRequestMapsSemanticFilters(t *testing.T) {
	req, err := buildAssetListRequest(assetListOptions{
		page:       2,
		pageSize:   20,
		id:         7,
		name:       "数据库",
		ip:         "192.0.2.0/24",
		mac:        "00:11:22:33:44:55",
		assetType:  "server",
		importance: "important",
		tagID:      "1,2",
		groupID:    3,
	})
	if err != nil {
		t.Fatalf("buildAssetListRequest returned error: %v", err)
	}

	if req["offset"] != int64(20) || req["count"] != int64(20) {
		t.Fatalf("pagination mismatch: %#v", req)
	}
	if req["id"] != uint(7) || req["name"] != "数据库" || req["ip"] != "192.0.2.0/24" {
		t.Fatalf("basic filters mismatch: %#v", req)
	}
	if req["importance"] != uint(1) {
		t.Fatalf("importance = %#v", req["importance"])
	}
	tagIDs := req["tag_id"].([]uint)
	if len(tagIDs) != 2 || tagIDs[1] != 2 {
		t.Fatalf("tag_id = %#v", tagIDs)
	}
	if req["group_id"] != uint(3) {
		t.Fatalf("group_id = %#v", req["group_id"])
	}
}

func TestBuildAssetGroupTreeRequestMapsSemanticFilters(t *testing.T) {
	req := buildAssetGroupTreeRequest(assetGroupTreeOptions{
		id:        3,
		depth:     4,
		withAsset: true,
	})

	if req["id"] != uint(3) || req["depth"] != 4 || req["with_asset"] != true {
		t.Fatalf("request = %#v", req)
	}
}

func TestBuildAssetDownloadTemplateRequest(t *testing.T) {
	req := buildAssetDownloadTemplateRequest(assetDownloadTemplateOptions{withExample: true})

	if req["with_data"] != false || req["with_example"] != true {
		t.Fatalf("request = %#v", req)
	}
	ids, ok := req["id_list"].([]int64)
	if !ok || len(ids) != 0 {
		t.Fatalf("id_list = %#v", req["id_list"])
	}
}

func TestBuildAssetExportRequest(t *testing.T) {
	req, err := buildAssetExportRequest(assetExportOptions{idList: "3,7"})
	if err != nil {
		t.Fatalf("buildAssetExportRequest returned error: %v", err)
	}

	if req["with_data"] != true || req["with_example"] != false {
		t.Fatalf("request = %#v", req)
	}
	ids := req["id_list"].([]int64)
	if len(ids) != 2 || ids[0] != 3 || ids[1] != 7 {
		t.Fatalf("id_list = %#v", ids)
	}
}

func TestBuildAssetCreateRequestMapsFields(t *testing.T) {
	req, err := buildAssetCreateRequest(assetCreateOptions{
		name:       "core-db",
		ip:         "192.0.2.10",
		contact:    "secops",
		importance: "important",
		remark:     "prod",
		assetType:  "server",
		location:   "beijing",
		tagID:      "3,7",
		groupID:    2,
		ipMacJSON:  `[{"ip":["192.0.2.10"],"mac":["00:11:22:33:44:55"]}]`,
	})
	if err != nil {
		t.Fatalf("buildAssetCreateRequest returned error: %v", err)
	}

	if req["name"] != "core-db" || req["ip"] != "192.0.2.10" || req["importance"] != uint(1) {
		t.Fatalf("basic request = %#v", req)
	}
	tags := req["tag_id_list"].([]uint)
	if len(tags) != 2 || tags[0] != 3 || tags[1] != 7 {
		t.Fatalf("tag_id_list = %#v", tags)
	}
	ipMac := req["ip_mac"].([]assetCreateIPMac)
	if len(ipMac) != 1 || ipMac[0].IP[0] != "192.0.2.10" || ipMac[0].Mac[0] != "00:11:22:33:44:55" {
		t.Fatalf("ip_mac = %#v", ipMac)
	}
}

func TestBuildAssetUpdateRequestMapsFields(t *testing.T) {
	req, err := buildAssetUpdateRequest(assetUpdateOptions{
		id:         "9",
		name:       "core-db-new",
		ip:         "192.0.2.11",
		contact:    "secops",
		importance: "normal",
		groupID:    2,
	})
	if err != nil {
		t.Fatalf("buildAssetUpdateRequest returned error: %v", err)
	}

	if req["id"] != uint(9) || req["name"] != "core-db-new" || req["ip"] != "192.0.2.11" {
		t.Fatalf("request = %#v", req)
	}
	if req["importance"] != uint(2) || req["group_id"] != uint(2) {
		t.Fatalf("request = %#v", req)
	}
}

func TestBuildAssetDeleteRequestMapsIDs(t *testing.T) {
	req, err := buildAssetDeleteRequest(assetDeleteOptions{idList: "9, 10"})
	if err != nil {
		t.Fatalf("buildAssetDeleteRequest returned error: %v", err)
	}
	ids, ok := req["ids"].([]uint)
	if !ok {
		t.Fatalf("ids type = %T", req["ids"])
	}
	if len(ids) != 2 || ids[0] != 9 || ids[1] != 10 {
		t.Fatalf("ids = %#v", ids)
	}
}

func TestBuildAssetBatchMaintainRequestMapsFields(t *testing.T) {
	req, err := buildAssetBatchMaintainRequest(assetBatchMaintainOptions{
		idList:   "9, 10",
		contact:  "secops",
		location: "beijing",
		remark:   "prod",
		groupID:  2,
	})
	if err != nil {
		t.Fatalf("buildAssetBatchMaintainRequest returned error: %v", err)
	}
	ids, ok := req["ids"].([]uint)
	if !ok {
		t.Fatalf("ids type = %T", req["ids"])
	}
	if len(ids) != 2 || ids[0] != 9 || ids[1] != 10 {
		t.Fatalf("ids = %#v", ids)
	}
	if req["contact"] != "secops" || req["location"] != "beijing" || req["remark"] != "prod" || req["group_id"] != uint(2) {
		t.Fatalf("request = %#v", req)
	}
}

func TestBuildAssetBatchTagRequestMapsFields(t *testing.T) {
	req, err := buildAssetBatchTagRequest(assetBatchTagOptions{
		idList: "9, 10",
		tagID:  "3,7",
	})
	if err != nil {
		t.Fatalf("buildAssetBatchTagRequest returned error: %v", err)
	}
	ids := req["ids"].([]uint)
	tagIDs := req["tag_ids"].([]uint)
	if len(ids) != 2 || ids[0] != 9 || ids[1] != 10 {
		t.Fatalf("ids = %#v", ids)
	}
	if len(tagIDs) != 2 || tagIDs[0] != 3 || tagIDs[1] != 7 {
		t.Fatalf("tag_ids = %#v", tagIDs)
	}
}

func TestBuildAssetGroupCreateRequestMapsFields(t *testing.T) {
	req, err := buildAssetGroupCreateRequest(assetGroupCreateOptions{name: "生产区", parentID: 2})
	if err != nil {
		t.Fatalf("buildAssetGroupCreateRequest returned error: %v", err)
	}
	if req["name"] != "生产区" || req["pid"] != uint(2) {
		t.Fatalf("request = %#v", req)
	}
}

func TestBuildAssetGroupRenameRequestMapsFields(t *testing.T) {
	req, err := buildAssetGroupRenameRequest(assetGroupRenameOptions{id: "3", name: "核心区"})
	if err != nil {
		t.Fatalf("buildAssetGroupRenameRequest returned error: %v", err)
	}
	if req["id"] != uint(3) || req["name"] != "核心区" {
		t.Fatalf("request = %#v", req)
	}
}

func TestBuildAssetGroupDeleteRequestRejectsRootGroup(t *testing.T) {
	if _, err := buildAssetGroupDeleteRequest(assetGroupDeleteOptions{idList: "1"}); err == nil {
		t.Fatal("expected root group delete to be rejected")
	}
	req, err := buildAssetGroupDeleteRequest(assetGroupDeleteOptions{idList: "3,4"})
	if err != nil {
		t.Fatalf("buildAssetGroupDeleteRequest returned error: %v", err)
	}
	ids := req["ids"].([]uint)
	if len(ids) != 2 || ids[0] != 3 || ids[1] != 4 {
		t.Fatalf("ids = %#v", ids)
	}
}

func TestBuildAssetTreeMoveRequestMapsFields(t *testing.T) {
	req, err := buildAssetTreeMoveRequest(assetTreeMoveOptions{
		id:       "9",
		nodeType: "asset",
		prevID:   "3",
		prevType: "group",
		topLayer: true,
	})
	if err != nil {
		t.Fatalf("buildAssetTreeMoveRequest returned error: %v", err)
	}
	if req["id"] != uint(9) || req["type"] != 2 || req["prev_id"] != uint(3) || req["prev_type"] != 1 || req["top_layer"] != true {
		t.Fatalf("request = %#v", req)
	}
}

func TestBuildAssetImportRequestMapsFileMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "assets.xlsx")
	if err := os.WriteFile(path, []byte("xlsx-bytes"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	req, err := buildAssetImportRequest(assetImportOptions{filePath: path})
	if err != nil {
		t.Fatalf("buildAssetImportRequest returned error: %v", err)
	}
	if req["file_path"] != path || req["file_name"] != "assets.xlsx" || req["size_bytes"] != int64(10) {
		t.Fatalf("request = %#v", req)
	}
	if req["method"] != assetImportMethod {
		t.Fatalf("method = %#v", req["method"])
	}
}

func TestAssetListCommandCallsGetAssetList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AssetService.GetAssetList" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["count"] != float64(5) || params["offset"] != float64(0) || params["ip"] != "192.0.2.10" {
			t.Fatalf("params = %#v", params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"data":[{"id":3,"name":"core-db","ip":"192.0.2.10","importance":1,"asset_type":"server","contact":"secops","group_info":{"id":1,"name":"default","path":[1]},"create_time":"2026-07-17 10:00:00","update_time":"2026-07-17 10:00:00"}],
				"total": 1
			}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "list",
		"--page-size", "5",
		"--ip", "192.0.2.10",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询资产列表" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["total"] != float64(1) || data["current_count"] != float64(1) {
		t.Fatalf("data counts = %#v", data)
	}
	assets := data["assets"].([]any)
	asset := assets[0].(map[string]any)
	if asset["name"] != "core-db" || asset["ip"] != "192.0.2.10" {
		t.Fatalf("asset summary mismatch: %#v", asset)
	}
}

func TestAssetGroupTreeCommandCallsSearchGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AssetService.SearchGroups" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["id"] != float64(1) || params["depth"] != float64(2) || params["with_asset"] != true {
			t.Fatalf("params = %#v", params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: json.RawMessage(`{
				"data":[
					{"id":1,"name":"全部资产","type":1,"count":2,"children":[
						{"id":3,"name":"生产区","type":1,"count":1},
						{"id":8,"name":"192.0.2.18","type":2,"count":0}
					]}
				]
			}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "group-tree",
		"--id", "1",
		"--depth", "2",
		"--with-asset",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查询资产组树" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["root_id"] != float64(1) || data["depth"] != float64(2) || data["with_asset"] != true {
		t.Fatalf("query data = %#v", data)
	}
	roots := data["groups"].([]any)
	root := roots[0].(map[string]any)
	if root["name"] != "全部资产" || root["count"] != float64(2) {
		t.Fatalf("root summary = %#v", root)
	}
	children := root["children"].([]any)
	child := children[0].(map[string]any)
	if child["id"] != float64(3) || child["path"] != "全部资产 / 生产区" {
		t.Fatalf("child summary = %#v", child)
	}
}

func TestAssetDownloadTemplateCommandDownloadsFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "asset-template.xlsx")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAssetDownloadRequest(t, r, false, true, nil)
		w.Header().Set("Content-Disposition", `attachment; filename="asset-template.xlsx";`)
		_, _ = io.WriteString(w, "template-bytes")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "download-template",
		"--with-example",
		"--output", outputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	rawFile, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(rawFile) != "template-bytes" {
		t.Fatalf("file content = %q", string(rawFile))
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "下载资产导入模板" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["file_path"] != outputPath || data["size_bytes"] != float64(len("template-bytes")) {
		t.Fatalf("data = %#v", data)
	}
}

func TestAssetExportCommandDownloadsFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "asset-export.xlsx")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAssetDownloadRequest(t, r, true, false, []int64{3, 7})
		w.Header().Set("Content-Disposition", `attachment; filename="asset-export.xlsx";`)
		_, _ = io.WriteString(w, "export-bytes")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "export",
		"--id-list", "3,7",
		"--output", outputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	rawFile, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(rawFile) != "export-bytes" {
		t.Fatalf("file content = %q", string(rawFile))
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "导出资产" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["file_path"] != outputPath || data["export_scope"] != "selected" {
		t.Fatalf("data = %#v", data)
	}
}

func TestAssetCreatePreviewDoesNotCallRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("preview must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "create",
		"--name", "core-db",
		"--ip", "192.0.2.10",
		"--preview",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called during preview")
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "新增资产预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmed"] != false || data["confirmation_token"] != "CONFIRM_ASSET_CREATE" {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestAssetCreateRequiresExactConfirmBeforeRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "create",
		"--name", "core-db",
		"--ip", "192.0.2.10",
		"--confirm", "confirm_asset_create",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "ASSET_CREATE_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestAssetCreateConfirmedCallsCreateAsset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AssetService.CreateAsset" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["name"] != "core-db" || params["ip"] != "192.0.2.10" || params["group_id"] != float64(2) {
			t.Fatalf("params = %#v", params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"id": 9}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "create",
		"--name", "core-db",
		"--ip", "192.0.2.10",
		"--group-id", "2",
		"--confirm", "CONFIRM_ASSET_CREATE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "新增资产" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	object := data["object"].(map[string]any)
	if object["id"] != float64(9) || object["ip"] != "192.0.2.10" {
		t.Fatalf("object = %#v", object)
	}
	audit := data["audit"].(map[string]any)
	if audit["action"] != "create" || audit["environment"] != server.URL {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestAssetDeletePreviewReadsBeforeAndDoesNotCallDelete(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		if req.Method != "AssetService.GetAssetInfo" {
			t.Fatalf("preview method = %q", req.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","ip":"192.0.2.10","importance":1,"group_id":1}}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "delete",
		"--id-list", "9",
		"--preview",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 1 || methods[0] != "AssetService.GetAssetInfo" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "删除资产预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmation_token"] != "CONFIRM_ASSET_DELETE" {
		t.Fatalf("preview data = %#v", data)
	}
	change := data["change_summary"].(map[string]any)
	before := change["before"].([]any)
	if len(before) != 1 {
		t.Fatalf("before = %#v", before)
	}
}

func TestAssetDeleteRequiresExactConfirmBeforeRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "delete",
		"--id-list", "9",
		"--confirm", "confirm_asset_delete",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "ASSET_DELETE_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestAssetDeleteConfirmedCallsDeleteAsset(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch req.Method {
		case "AssetService.GetAssetInfo":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","ip":"192.0.2.10","importance":1,"group_id":1}}`),
			})
		case "AssetService.DeleteAsset":
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			ids := params["ids"].([]any)
			if len(ids) != 1 || ids[0] != float64(9) {
				t.Fatalf("params = %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: json.RawMessage(`{}`)})
		default:
			t.Fatalf("unexpected method = %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "delete",
		"--id-list", "9",
		"--confirm", "CONFIRM_ASSET_DELETE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 2 || methods[0] != "AssetService.GetAssetInfo" || methods[1] != "AssetService.DeleteAsset" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "删除资产" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	audit := data["audit"].(map[string]any)
	if audit["action"] != "delete" || audit["environment"] != server.URL {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestAssetBatchMaintainPreviewReadsBeforeAndDoesNotCallUpdate(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		if req.Method != "AssetService.GetAssetInfo" {
			t.Fatalf("preview method = %q", req.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","ip":"192.0.2.10","contact":"old","group_id":1}}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "batch-maintain",
		"--id-list", "9",
		"--contact", "secops",
		"--preview",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 1 || methods[0] != "AssetService.GetAssetInfo" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "批量维护资产预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmation_token"] != "CONFIRM_ASSET_BATCH_MAINTAIN" {
		t.Fatalf("preview data = %#v", data)
	}
	change := data["change_summary"].(map[string]any)
	after := change["after"].(map[string]any)
	if after["contact"] != "secops" {
		t.Fatalf("change_summary = %#v", change)
	}
}

func TestAssetBatchMaintainRequiresExactConfirmBeforeRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "batch-maintain",
		"--id-list", "9",
		"--contact", "secops",
		"--confirm", "confirm_asset_batch_maintain",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "ASSET_BATCH_MAINTAIN_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestAssetBatchMaintainConfirmedCallsUpdateAssetBatch(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch req.Method {
		case "AssetService.GetAssetInfo":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","ip":"192.0.2.10","contact":"old","group_id":1}}`),
			})
		case "AssetService.UpdateAssetBatch":
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			ids := params["ids"].([]any)
			if len(ids) != 1 || ids[0] != float64(9) || params["contact"] != "secops" || params["group_id"] != float64(2) {
				t.Fatalf("params = %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: json.RawMessage(`{}`)})
		default:
			t.Fatalf("unexpected method = %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "batch-maintain",
		"--id-list", "9",
		"--contact", "secops",
		"--group-id", "2",
		"--confirm", "CONFIRM_ASSET_BATCH_MAINTAIN",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 2 || methods[0] != "AssetService.GetAssetInfo" || methods[1] != "AssetService.UpdateAssetBatch" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "批量维护资产" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	audit := data["audit"].(map[string]any)
	if audit["action"] != "batch_maintain" || audit["environment"] != server.URL {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestAssetBatchTagPreviewReadsBeforeAndDoesNotCallUpdate(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		if req.Method != "AssetService.GetAssetInfo" {
			t.Fatalf("preview method = %q", req.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","tags":[{"id":1,"name":"old"}]}}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "batch-tag",
		"--id-list", "9",
		"--tag-id", "3,7",
		"--preview",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 1 || methods[0] != "AssetService.GetAssetInfo" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "批量维护资产标签预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmation_token"] != "CONFIRM_ASSET_BATCH_TAG" {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestAssetBatchTagRequiresExactConfirmBeforeRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "batch-tag",
		"--id-list", "9",
		"--tag-id", "3",
		"--confirm", "confirm_asset_batch_tag",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "ASSET_BATCH_TAG_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestAssetBatchTagConfirmedCallsUpdateAssetTagBatch(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch req.Method {
		case "AssetService.GetAssetInfo":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","tags":[{"id":1,"name":"old"}]}}`),
			})
		case "AssetService.UpdateAssetTagBatch":
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			ids := params["ids"].([]any)
			tagIDs := params["tag_ids"].([]any)
			if len(ids) != 1 || ids[0] != float64(9) || len(tagIDs) != 2 || tagIDs[1] != float64(7) {
				t.Fatalf("params = %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: json.RawMessage(`{}`)})
		default:
			t.Fatalf("unexpected method = %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "batch-tag",
		"--id-list", "9",
		"--tag-id", "3,7",
		"--confirm", "CONFIRM_ASSET_BATCH_TAG",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 2 || methods[0] != "AssetService.GetAssetInfo" || methods[1] != "AssetService.UpdateAssetTagBatch" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "批量维护资产标签" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	audit := data["audit"].(map[string]any)
	if audit["action"] != "batch_tag" || audit["environment"] != server.URL {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestAssetGroupCreatePreviewReadsParentAndDoesNotCallCreate(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		if req.Method != "AssetService.SearchGroups" {
			t.Fatalf("preview method = %q", req.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":[{"id":2,"name":"生产","type":1,"count":3}]}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "group-create",
		"--name", "核心区",
		"--parent-id", "2",
		"--preview",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 1 || methods[0] != "AssetService.SearchGroups" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "创建资产组预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmation_token"] != "CONFIRM_ASSET_GROUP_CREATE" {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestAssetGroupRenameConfirmedCallsUpdateGroup(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch req.Method {
		case "AssetService.SearchGroups":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":[{"id":3,"name":"旧名称","type":1,"count":2}]}`),
			})
		case "AssetService.UpdateGroup":
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			if params["id"] != float64(3) || params["name"] != "核心区" {
				t.Fatalf("params = %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: json.RawMessage(`{}`)})
		default:
			t.Fatalf("unexpected method = %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "group-rename",
		"--id", "3",
		"--name", "核心区",
		"--confirm", "CONFIRM_ASSET_GROUP_RENAME",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 2 || methods[0] != "AssetService.SearchGroups" || methods[1] != "AssetService.UpdateGroup" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "重命名资产组" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	audit := data["audit"].(map[string]any)
	if audit["action"] != "group_rename" || audit["environment"] != server.URL {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestAssetGroupDeleteConfirmedCallsDeleteGroup(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch req.Method {
		case "AssetService.SearchGroups":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":[{"id":3,"name":"待删除","type":1,"count":2}]}`),
			})
		case "AssetService.DeleteGroup":
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			ids := params["ids"].([]any)
			if len(ids) != 1 || ids[0] != float64(3) {
				t.Fatalf("params = %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: json.RawMessage(`{}`)})
		default:
			t.Fatalf("unexpected method = %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "group-delete",
		"--id-list", "3",
		"--confirm", "CONFIRM_ASSET_GROUP_DELETE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 2 || methods[0] != "AssetService.SearchGroups" || methods[1] != "AssetService.DeleteGroup" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "删除资产组" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	audit := data["audit"].(map[string]any)
	if audit["action"] != "group_delete" || audit["environment"] != server.URL {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestAssetTreeMovePreviewReadsSourceAndPrevAndDoesNotCallMove(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch req.Method {
		case "AssetService.GetAssetInfo":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","ip":"192.0.2.10","group_id":1}}`),
			})
		case "AssetService.SearchGroups":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":[{"id":3,"name":"生产区","type":1,"count":2}]}`),
			})
		default:
			t.Fatalf("unexpected preview method = %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "tree-move",
		"--id", "9",
		"--type", "asset",
		"--prev-id", "3",
		"--prev-type", "group",
		"--top-layer",
		"--preview",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 2 || methods[0] != "AssetService.GetAssetInfo" || methods[1] != "AssetService.SearchGroups" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "调整资产树层级预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmation_token"] != "CONFIRM_ASSET_TREE_MOVE" {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestAssetTreeMoveRequiresExactConfirmBeforeRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "tree-move",
		"--id", "9",
		"--type", "asset",
		"--prev-id", "3",
		"--prev-type", "group",
		"--confirm", "confirm_asset_tree_move",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "ASSET_TREE_MOVE_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestAssetTreeMoveConfirmedCallsUpdateAssetTree(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch req.Method {
		case "AssetService.GetAssetInfo":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","ip":"192.0.2.10","group_id":1}}`),
			})
		case "AssetService.SearchGroups":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":[{"id":3,"name":"生产区","type":1,"count":2}]}`),
			})
		case "AssetService.UpdateAssetTree":
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			if params["id"] != float64(9) || params["type"] != float64(2) || params["prev_id"] != float64(3) || params["prev_type"] != float64(1) || params["top_layer"] != true {
				t.Fatalf("params = %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: json.RawMessage(`{}`)})
		default:
			t.Fatalf("unexpected method = %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "tree-move",
		"--id", "9",
		"--type", "asset",
		"--prev-id", "3",
		"--prev-type", "group",
		"--top-layer",
		"--confirm", "CONFIRM_ASSET_TREE_MOVE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 3 || methods[2] != "AssetService.UpdateAssetTree" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "调整资产树层级" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	audit := data["audit"].(map[string]any)
	if audit["action"] != "tree_move" || audit["environment"] != server.URL {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestAssetImportPreviewDoesNotUpload(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("preview must not upload")
	}))
	defer server.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "assets.xlsx")
	if err := os.WriteFile(path, []byte("xlsx-bytes"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "import",
		"--file", path,
		"--preview",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend upload was called during preview")
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "导入资产预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmation_token"] != "CONFIRM_ASSET_IMPORT" {
		t.Fatalf("preview data = %#v", data)
	}
}

func TestAssetImportRequiresExactConfirmBeforeUpload(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not upload")
	}))
	defer server.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "assets.xlsx")
	if err := os.WriteFile(path, []byte("xlsx-bytes"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "import",
		"--file", path,
		"--confirm", "confirm_asset_import",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend upload was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "ASSET_IMPORT_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestAssetImportConfirmedUploadsFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/upload" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("id") != "AssetUploadService.UploadAsset" {
			t.Fatalf("id query = %q", r.URL.Query().Get("id"))
		}
		if r.Header.Get("Api-Token") != "token-123" {
			t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("MultipartReader: %v", err)
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Fatalf("NextPart: %v", err)
		}
		if part.FormName() != "file" || part.FileName() != "assets.xlsx" {
			t.Fatalf("part name=%q filename=%q", part.FormName(), part.FileName())
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      "upload",
			Result:  json.RawMessage(`{"record_count":{"record_total":2,"success_total":1,"fail_total":1,"cover_total":0,"failure_reason":[{"row":2,"reason":"invalid ip"}]}}`),
		})
	}))
	defer server.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "assets.xlsx")
	if err := os.WriteFile(path, []byte("xlsx-bytes"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "import",
		"--file", path,
		"--confirm", "CONFIRM_ASSET_IMPORT",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "导入资产" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	result := data["import_result"].(map[string]any)
	if result["record_total"] != float64(2) || result["fail_total"] != float64(1) {
		t.Fatalf("import_result = %#v", result)
	}
}

func TestAssetUpdatePreviewReadsBeforeAndDoesNotCallUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AssetService.GetAssetInfo" {
			t.Fatalf("preview method = %q", req.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","ip":"192.0.2.10","importance":1,"group_id":1}}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "update",
		"--id", "9",
		"--name", "core-db-new",
		"--ip", "192.0.2.11",
		"--preview",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "编辑资产预览" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["requires_confirmation"] != true || data["confirmation_token"] != "CONFIRM_ASSET_UPDATE" {
		t.Fatalf("preview data = %#v", data)
	}
	change := data["change_summary"].(map[string]any)
	before := change["before"].(map[string]any)
	after := change["after"].(map[string]any)
	if before["name"] != "core-db" || after["name"] != "core-db-new" {
		t.Fatalf("change_summary = %#v", change)
	}
}

func TestAssetUpdateRequiresExactConfirmBeforeRPC(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		t.Fatalf("invalid confirmation must not call backend RPC")
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "update",
		"--id", "9",
		"--name", "core-db-new",
		"--ip", "192.0.2.11",
		"--confirm", "confirm_asset_update",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if serverCalled {
		t.Fatal("backend RPC was called with invalid confirmation")
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Error.Code != "ASSET_UPDATE_CONFIRMATION_REQUIRED" {
		t.Fatalf("error = %#v", env.Error)
	}
}

func TestAssetUpdateConfirmedCallsUpdateAsset(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methods = append(methods, req.Method)
		switch req.Method {
		case "AssetService.GetAssetInfo":
			_ = json.NewEncoder(w).Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  json.RawMessage(`{"data":{"id":9,"name":"core-db","ip":"192.0.2.10","importance":1,"group_id":1}}`),
			})
		case "AssetService.UpdateAsset":
			raw, _ := json.Marshal(req.Params)
			var params map[string]any
			if err := json.Unmarshal(raw, &params); err != nil {
				t.Fatalf("decode params: %v", err)
			}
			if params["id"] != float64(9) || params["name"] != "core-db-new" || params["ip"] != "192.0.2.11" {
				t.Fatalf("params = %#v", params)
			}
			_ = json.NewEncoder(w).Encode(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: json.RawMessage(`{}`)})
		default:
			t.Fatalf("unexpected method = %q", req.Method)
		}
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "update",
		"--id", "9",
		"--name", "core-db-new",
		"--ip", "192.0.2.11",
		"--confirm", "CONFIRM_ASSET_UPDATE",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(methods) != 2 || methods[0] != "AssetService.GetAssetInfo" || methods[1] != "AssetService.UpdateAsset" {
		t.Fatalf("methods = %#v", methods)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "编辑资产" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["confirmed"] != true || data["result"] != "success" {
		t.Fatalf("data = %#v", data)
	}
	audit := data["audit"].(map[string]any)
	if audit["action"] != "update" || audit["environment"] != server.URL {
		t.Fatalf("audit = %#v", audit)
	}
}

func TestAssetDetailCommandCallsGetAssetInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AssetService.GetAssetInfo" {
			t.Fatalf("method = %q", req.Method)
		}
		raw, _ := json.Marshal(req.Params)
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			t.Fatalf("decode params: %v", err)
		}
		if params["id"] != float64(3) {
			t.Fatalf("id param = %#v", params)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":{"id":3,"name":"core-db","ip":"192.0.2.10","importance":1,"asset_type":"server"}}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer",
		"--url", server.URL,
		"--api-key", "token-123",
		"asset", "detail",
		"--id", "3",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env SuccessEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Task != "查看资产详情" {
		t.Fatalf("Task = %q", env.Task)
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["id"] != float64(3) {
		t.Fatalf("id = %#v", data["id"])
	}
	detail := data["detail"].(map[string]any)
	if detail["name"] != "core-db" {
		t.Fatalf("detail = %#v", detail)
	}
}

func TestAssetDetailRejectsEmptyBackendDetailAsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "AssetService.GetAssetInfo" {
			t.Fatalf("method = %q", req.Method)
		}
		_ = json.NewEncoder(w).Encode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"data":{"id":0}}`),
		})
	}))
	defer server.Close()

	var out bytes.Buffer
	cmd := NewRootCommand(RootOptions{Out: &out, ErrOut: &out})
	cmd.SetArgs([]string{
		"tanswer", "--url", server.URL, "--api-key", "token-123",
		"asset", "detail", "--id", "999999",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if env.Success || env.Error.Code != "ASSET_NOT_FOUND" || env.Error.Retryable {
		t.Fatalf("not-found error = %#v", env)
	}
}

func assertAssetDownloadRequest(t *testing.T, r *http.Request, withData bool, withExample bool, wantIDs []int64) {
	t.Helper()
	if r.URL.Path != "/api/download" {
		t.Fatalf("path = %q", r.URL.Path)
	}
	if r.Header.Get("Api-Token") != "token-123" {
		t.Fatalf("Api-Token header = %q", r.Header.Get("Api-Token"))
	}
	if r.URL.Query().Get("id") != "AssetDownloadServer.DownloadAssetTemplate" {
		t.Fatalf("id query = %q", r.URL.Query().Get("id"))
	}
	raw, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("query"))
	if err != nil {
		t.Fatalf("decode query: %v", err)
	}
	var query map[string]any
	if err := json.Unmarshal(raw, &query); err != nil {
		t.Fatalf("decode query json: %v", err)
	}
	if query["with_data"] != withData || query["with_example"] != withExample {
		t.Fatalf("query = %#v", query)
	}
	ids := query["id_list"].([]any)
	if len(ids) != len(wantIDs) {
		t.Fatalf("id_list = %#v", ids)
	}
	for i, id := range wantIDs {
		if ids[i] != float64(id) {
			t.Fatalf("id_list = %#v", ids)
		}
	}
}
