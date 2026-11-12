package tanswer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const assetDownloadMethod = "AssetDownloadServer.DownloadAssetTemplate"
const assetImportMethod = "AssetUploadService.UploadAsset"
const assetCreateConfirmToken = "CONFIRM_ASSET_CREATE"
const assetUpdateConfirmToken = "CONFIRM_ASSET_UPDATE"
const assetDeleteConfirmToken = "CONFIRM_ASSET_DELETE"
const assetBatchMaintainConfirmToken = "CONFIRM_ASSET_BATCH_MAINTAIN"
const assetBatchTagConfirmToken = "CONFIRM_ASSET_BATCH_TAG"
const assetGroupCreateConfirmToken = "CONFIRM_ASSET_GROUP_CREATE"
const assetGroupRenameConfirmToken = "CONFIRM_ASSET_GROUP_RENAME"
const assetGroupDeleteConfirmToken = "CONFIRM_ASSET_GROUP_DELETE"
const assetTreeMoveConfirmToken = "CONFIRM_ASSET_TREE_MOVE"
const assetImportConfirmToken = "CONFIRM_ASSET_IMPORT"

const assetTreeNodeTypeGroup = 1
const assetTreeNodeTypeAsset = 2

type assetListOptions struct {
	page       int
	pageSize   int
	id         uint
	name       string
	ip         string
	mac        string
	assetType  string
	importance string
	tagID      string
	groupID    uint
}

type assetDetailOptions struct {
	id string
}

type assetGroupTreeOptions struct {
	id        uint
	depth     int
	withAsset bool
}

type assetDownloadTemplateOptions struct {
	output      string
	withExample bool
}

type assetImportOptions struct {
	filePath string
	preview  bool
	confirm  string
}

type assetExportOptions struct {
	output string
	idList string
}

type assetCreateOptions struct {
	name       string
	ip         string
	contact    string
	importance string
	remark     string
	assetType  string
	location   string
	tagID      string
	groupID    uint
	ipMacJSON  string
	preview    bool
	confirm    string
}

type assetUpdateOptions struct {
	id         string
	name       string
	ip         string
	contact    string
	importance string
	remark     string
	assetType  string
	location   string
	tagID      string
	groupID    uint
	ipMacJSON  string
	preview    bool
	confirm    string
}

type assetDeleteOptions struct {
	idList  string
	preview bool
	confirm string
}

type assetBatchMaintainOptions struct {
	idList   string
	contact  string
	remark   string
	location string
	groupID  uint
	preview  bool
	confirm  string
}

type assetBatchTagOptions struct {
	idList  string
	tagID   string
	preview bool
	confirm string
}

type assetGroupCreateOptions struct {
	name     string
	parentID uint
	preview  bool
	confirm  string
}

type assetGroupRenameOptions struct {
	id      string
	name    string
	preview bool
	confirm string
}

type assetGroupDeleteOptions struct {
	idList  string
	preview bool
	confirm string
}

type assetTreeMoveOptions struct {
	id       string
	nodeType string
	prevID   string
	prevType string
	topLayer bool
	preview  bool
	confirm  string
}

type assetCreateIPMac struct {
	IP  []string `json:"ip"`
	Mac []string `json:"mac"`
}

type assetListRPCResult struct {
	Data  []map[string]any `json:"data"`
	Total int64            `json:"total"`
}

type assetDetailRPCResult struct {
	Data map[string]any `json:"data"`
}

type assetGroupTreeRPCResult struct {
	Data []assetGroupTreeNode `json:"data"`
}

type assetGroupTreeNode struct {
	Children []assetGroupTreeNode `json:"children,omitempty"`
	ID       uint                 `json:"id"`
	Name     string               `json:"name"`
	Type     int                  `json:"type"`
	Count    int64                `json:"count"`
}

func newAssetCommand(opts *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "资产配置语义命令",
		Long:  "资产配置语义命令。用于查询资产列表、资产详情、资产组树、下载资产导入模板和导出资产；资产风险能力不在当前版本范围内。",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newAssetListCommand(opts))
	cmd.AddCommand(newAssetDetailCommand(opts))
	cmd.AddCommand(newAssetGroupTreeCommand(opts))
	cmd.AddCommand(newAssetDownloadTemplateCommand(opts))
	cmd.AddCommand(newAssetExportCommand(opts))
	cmd.AddCommand(newAssetCreateCommand(opts))
	cmd.AddCommand(newAssetUpdateCommand(opts))
	cmd.AddCommand(newAssetDeleteCommand(opts))
	cmd.AddCommand(newAssetBatchMaintainCommand(opts))
	cmd.AddCommand(newAssetBatchTagCommand(opts))
	cmd.AddCommand(newAssetGroupCreateCommand(opts))
	cmd.AddCommand(newAssetGroupRenameCommand(opts))
	cmd.AddCommand(newAssetGroupDeleteCommand(opts))
	cmd.AddCommand(newAssetTreeMoveCommand(opts))
	cmd.AddCommand(newAssetImportCommand(opts))
	return cmd
}

func newAssetListCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "查询资产列表",
		Long: "查询资产列表，用于查看当前资产配置和筛选结果。支持按资产名称、IP、MAC、资产类型、资产等级、资产标签和资产组筛选，并稳定返回分页信息。\n\n" +
			"输出：实际筛选条件、total、page、page_size、current_count、has_more、assets。列表项保留产品资产档案摘要字段，如 id、name、ip、ip_mac、importance、tags、asset_type、group_info、contact、remark、location、source、create_time、update_time。",
		Example: "  chaitin-cli tanswer asset list --page-size 10\n" +
			"  chaitin-cli tanswer asset list --ip 192.0.2.10\n" +
			"  chaitin-cli tanswer asset list --importance important --asset-type server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if assetOpts.page < 1 {
				return writeAssetError(cmd, "查询资产列表", "chaitin-cli tanswer asset list", "INVALID_PAGE", "page must be greater than or equal to 1", false)
			}
			if assetOpts.pageSize < 1 || assetOpts.pageSize > 100 {
				return writeAssetError(cmd, "查询资产列表", "chaitin-cli tanswer asset list", "INVALID_PAGE_SIZE", "page-size must be between 1 and 100", false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			req, err := buildAssetListRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "查询资产列表", "chaitin-cli tanswer asset list", "INVALID_ASSET_FILTER", err.Error(), false)
			}
			client := NewClient(cfg)
			var result assetListRPCResult
			if err := client.CallRPC(cmd.Context(), "AssetService.GetAssetList", req, &result); err != nil {
				return writeAssetError(cmd, "查询资产列表", "chaitin-cli tanswer asset list", "ASSET_LIST_FAILED", err.Error(), true)
			}
			data := map[string]any{
				"total":         result.Total,
				"page":          assetOpts.page,
				"page_size":     assetOpts.pageSize,
				"current_count": len(result.Data),
				"has_more":      int64(assetOpts.page*assetOpts.pageSize) < result.Total,
				"assets":        summarizeAssets(result.Data),
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查询资产列表",
				Command: "chaitin-cli tanswer asset list",
				Query: map[string]any{
					"filters":   assetListFilters(assetOpts),
					"page":      assetOpts.page,
					"page_size": assetOpts.pageSize,
				},
				Data: data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	addAssetListFlags(cmd, &assetOpts)
	return cmd
}

func newAssetDetailCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetDetailOptions
	cmd := &cobra.Command{
		Use:   "detail",
		Short: "查看资产详情",
		Long: "查看资产详情，用于确认资产归属和基础信息。该命令返回产品资产档案详情字段，包括资产名称、IP、MAC、资产类型、资产等级、资产组、资产标签、负责人、地理位置、备注、来源、添加时间和最后编辑时间。\n\n" +
			"输出：id、detail。detail 中保留后端资产详情字段。",
		Example: "  chaitin-cli tanswer asset detail --id '<asset_id>'",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseRequiredAssetID(assetOpts.id)
			if err != nil {
				return writeAssetError(cmd, "查看资产详情", "chaitin-cli tanswer asset detail", "MISSING_ASSET_ID", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result assetDetailRPCResult
			if err := client.CallRPC(cmd.Context(), "AssetService.GetAssetInfo", map[string]any{"id": id}, &result); err != nil {
				return writeAssetError(cmd, "查看资产详情", "chaitin-cli tanswer asset detail", "ASSET_DETAIL_FAILED", err.Error(), true)
			}
			if assetDetailIsMissing(result.Data) {
				return writeAssetError(cmd, "查看资产详情", "chaitin-cli tanswer asset detail", "ASSET_NOT_FOUND", fmt.Sprintf("asset %d was not found", id), false)
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查看资产详情",
				Command: "chaitin-cli tanswer asset detail",
				Query:   map[string]any{"id": id},
				Data: map[string]any{
					"id":     id,
					"detail": result.Data,
				},
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.id, "id", "", "asset id from asset list")
	return cmd
}

func newAssetGroupTreeCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetGroupTreeOptions
	cmd := &cobra.Command{
		Use:   "group-tree",
		Short: "查询资产组树",
		Long: "查询资产组树，用于查看当前资产组层级，并在资产查询、导入或批量维护前选择资产组 ID。默认从根资产组开始查询，默认不展开资产节点，避免一次输出过多数据。\n\n" +
			"输出：root_id、depth、with_asset、target、groups。每个节点包含 id、name、type、type_label、count、path、children。",
		Example: "  chaitin-cli tanswer asset group-tree\n" +
			"  chaitin-cli tanswer asset group-tree --id 3 --depth 2\n" +
			"  chaitin-cli tanswer asset group-tree --with-asset --depth 1",
		RunE: func(cmd *cobra.Command, args []string) error {
			if assetOpts.depth < 1 || assetOpts.depth > 100 {
				return writeAssetError(cmd, "查询资产组树", "chaitin-cli tanswer asset group-tree", "INVALID_DEPTH", "depth must be between 1 and 100", false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			req := buildAssetGroupTreeRequest(assetOpts)
			client := NewClient(cfg)
			var result assetGroupTreeRPCResult
			if err := client.CallRPC(cmd.Context(), "AssetService.SearchGroups", req, &result); err != nil {
				return writeAssetError(cmd, "查询资产组树", "chaitin-cli tanswer asset group-tree", "ASSET_GROUP_TREE_FAILED", err.Error(), true)
			}
			groups := summarizeAssetGroupTree(result.Data, nil)
			data := map[string]any{
				"root_id":    assetOpts.id,
				"depth":      assetOpts.depth,
				"with_asset": assetOpts.withAsset,
				"target":     firstAssetGroupSummary(groups),
				"groups":     groups,
			}
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "查询资产组树",
				Command: "chaitin-cli tanswer asset group-tree",
				Query:   req,
				Data:    data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().UintVar(&assetOpts.id, "id", 1, "资产组 ID，默认根资产组 1")
	cmd.Flags().IntVar(&assetOpts.depth, "depth", 3, "树查询深度，1-100")
	cmd.Flags().BoolVar(&assetOpts.withAsset, "with-asset", false, "是否展开资产节点；默认只返回资产组")
	return cmd
}

func newAssetDownloadTemplateCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetDownloadTemplateOptions
	cmd := &cobra.Command{
		Use:   "download-template",
		Short: "下载资产导入模板",
		Long: "下载资产导入模板，用于获取产品当前资产导入 Excel 模板，供后续按模板填写资产后再执行导入。该命令只下载模板文件，不导入、不创建、不修改资产。\n\n" +
			"输出：file_name、file_path、size_bytes、status_code、method、download_query。",
		Example: "  chaitin-cli tanswer asset download-template --output ./asset-template.xlsx\n" +
			"  chaitin-cli tanswer asset download-template --with-example --output ./asset-template-example.xlsx",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			query := buildAssetDownloadTemplateRequest(assetOpts)
			return downloadAssetFile(cmd, cfg, "下载资产导入模板", "chaitin-cli tanswer asset download-template", query, assetOpts.output, "asset-template.xlsx", "")
		},
	}
	cmd.Flags().StringVar(&assetOpts.output, "output", "", "output file path; defaults to downloaded filename or asset-template.xlsx")
	cmd.Flags().BoolVar(&assetOpts.withExample, "with-example", false, "download template with example rows")
	return cmd
}

func newAssetImportCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetImportOptions
	cmd := &cobra.Command{
		Use:   "import",
		Short: "导入资产",
		Long: "导入资产，用于上传产品资产导入模板文件并批量初始化或更新资产。该命令是高影响写操作：预览阶段只读取本地文件元信息，不上传文件；必须使用 --confirm CONFIRM_ASSET_IMPORT 才会调用后端上传接口。\n\n" +
			"执行输出：confirmed、result、object、audit、import_result。",
		Example: "  chaitin-cli tanswer asset import --file ./assets.xlsx --preview\n" +
			"  chaitin-cli tanswer asset import --file ./assets.xlsx --confirm CONFIRM_ASSET_IMPORT",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetImportRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "导入资产", "chaitin-cli tanswer asset import", "INVALID_ASSET_IMPORT_REQUEST", err.Error(), false)
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "导入资产预览",
				Command:       "chaitin-cli tanswer asset import",
				OperationType: "asset_import",
				RiskLevel:     "write_high",
				Target:        map[string]any{"file_name": req["file_name"], "file_path": req["file_path"]},
				ChangeSummary: map[string]any{"before": nil, "after": req},
				Impact:        map[string]any{"file_size_bytes": req["size_bytes"]},
				RiskWarnings: []string{
					"将上传资产导入文件并批量创建或更新资产配置。",
					"导入结果由后端按模板内容返回，可能包含行级失败原因。",
				},
				ConfirmToken: assetImportConfirmToken,
			})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "导入资产预览", Command: "chaitin-cli tanswer asset import", Query: req, Data: preview})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := ValidateWriteConfirmation(assetOpts.confirm, assetImportConfirmToken); err != nil {
				return writeAssetError(cmd, "导入资产", "chaitin-cli tanswer asset import", "ASSET_IMPORT_CONFIRMATION_REQUIRED", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			file, err := os.Open(req["file_path"].(string))
			if err != nil {
				return writeAssetError(cmd, "导入资产", "chaitin-cli tanswer asset import", "ASSET_IMPORT_FILE_OPEN_FAILED", err.Error(), false)
			}
			defer file.Close()
			client := NewClient(cfg)
			resp, err := client.UploadFile(cmd.Context(), assetImportMethod, req["file_name"].(string), file)
			if err != nil {
				return writeAssetError(cmd, "导入资产", "chaitin-cli tanswer asset import", "ASSET_IMPORT_FAILED", err.Error(), true)
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return writeAssetError(cmd, "导入资产", "chaitin-cli tanswer asset import", "ASSET_IMPORT_FAILED", fmt.Sprintf("upload returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(resp.Body))), true)
			}
			importResult, err := parseAssetImportResult(resp.Body)
			if err != nil {
				return writeAssetError(cmd, "导入资产", "chaitin-cli tanswer asset import", "ASSET_IMPORT_RESPONSE_INVALID", err.Error(), true)
			}
			object := map[string]any{"file_name": req["file_name"], "file_path": req["file_path"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "asset_import", Object: object, Action: "import", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: map[string]any{"before": nil, "after": req}, Result: "success"})
			data["import_result"] = importResult
			raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "导入资产", Command: "chaitin-cli tanswer asset import", Query: req, Data: data})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.filePath, "file", "", "asset import template file path, required")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without uploading")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to upload")
	return cmd
}

func newAssetExportCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetExportOptions
	cmd := &cobra.Command{
		Use:   "export",
		Short: "导出资产",
		Long: "导出资产，用于备份或离线编辑当前资产配置。未指定 --id-list 时导出全部资产；指定 --id-list 时只导出选中资产。该命令只下载导出文件，不修改资产配置。\n\n" +
			"输出：file_name、file_path、size_bytes、status_code、method、download_query、export_scope。",
		Example: "  chaitin-cli tanswer asset export --output ./asset-export.xlsx\n" +
			"  chaitin-cli tanswer asset export --id-list 3,7 --output ./selected-assets.xlsx",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			query, err := buildAssetExportRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "导出资产", "chaitin-cli tanswer asset export", "INVALID_ID_LIST", err.Error(), false)
			}
			scope := "all"
			if ids, ok := query["id_list"].([]int64); ok && len(ids) > 0 {
				scope = "selected"
			}
			return downloadAssetFile(cmd, cfg, "导出资产", "chaitin-cli tanswer asset export", query, assetOpts.output, "asset-export.xlsx", scope)
		},
	}
	cmd.Flags().StringVar(&assetOpts.output, "output", "", "output file path; defaults to downloaded filename or asset-export.xlsx")
	cmd.Flags().StringVar(&assetOpts.idList, "id-list", "", "comma-separated asset IDs to export; empty exports all assets")
	return cmd
}

func newAssetCreateCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetCreateOptions
	cmd := &cobra.Command{
		Use:   "create",
		Short: "新增资产",
		Long: "新增资产，将新资产纳入识别和运营范围。该命令是高影响写操作：默认只返回变更预览；必须在审阅预览后使用 --confirm CONFIRM_ASSET_CREATE 才会调用后端写入接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer asset create --name core-db --ip 192.0.2.10 --preview\n" +
			"  chaitin-cli tanswer asset create --name core-db --ip 192.0.2.10 --group-id 2 --confirm CONFIRM_ASSET_CREATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetCreateRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "新增资产", "chaitin-cli tanswer asset create", "INVALID_ASSET_CREATE_REQUEST", err.Error(), false)
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "新增资产预览",
				Command:       "chaitin-cli tanswer asset create",
				OperationType: "asset_create",
				RiskLevel:     "write_high",
				Target: map[string]any{
					"name": req["name"],
					"ip":   req["ip"],
				},
				ChangeSummary: req,
				Impact: map[string]any{
					"asset_count": 1,
					"group_id":    req["group_id"],
				},
				RiskWarnings: []string{
					"将新增资产配置并触发资产配置版本更新。",
					"如果 IP 与已有资产冲突，后端会拒绝写入。",
				},
				ConfirmToken: assetCreateConfirmToken,
			})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{
					Success: true,
					Task:    "新增资产预览",
					Command: "chaitin-cli tanswer asset create",
					Query:   req,
					Data:    preview,
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := ValidateWriteConfirmation(assetOpts.confirm, assetCreateConfirmToken); err != nil {
				return writeAssetError(cmd, "新增资产", "chaitin-cli tanswer asset create", "ASSET_CREATE_CONFIRMATION_REQUIRED", err.Error(), false)
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			var result struct {
				ID uint `json:"id"`
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.CreateAsset", req, &result); err != nil {
				return writeAssetError(cmd, "新增资产", "chaitin-cli tanswer asset create", "ASSET_CREATE_FAILED", err.Error(), true)
			}
			object := map[string]any{
				"id":         result.ID,
				"name":       req["name"],
				"ip":         req["ip"],
				"group_id":   req["group_id"],
				"importance": req["importance"],
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "asset_create",
				Object:        object,
				Action:        "create",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter: map[string]any{
					"before": nil,
					"after":  req,
				},
				Result: "success",
			})
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "新增资产",
				Command: "chaitin-cli tanswer asset create",
				Query:   req,
				Data:    data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.name, "name", "", "asset name, required")
	cmd.Flags().StringVar(&assetOpts.ip, "ip", "", "asset IP list, comma separated, required")
	cmd.Flags().StringVar(&assetOpts.contact, "contact", "", "asset owner/contact")
	cmd.Flags().StringVar(&assetOpts.importance, "importance", "normal", "asset importance: important, normal, 重点, 普通, 1, or 2")
	cmd.Flags().StringVar(&assetOpts.remark, "remark", "", "asset remark")
	cmd.Flags().StringVar(&assetOpts.assetType, "asset-type", "", "asset type")
	cmd.Flags().StringVar(&assetOpts.location, "location", "", "asset location")
	cmd.Flags().StringVar(&assetOpts.tagID, "tag-id", "", "asset tag IDs, comma separated")
	cmd.Flags().UintVar(&assetOpts.groupID, "group-id", 1, "asset group ID")
	cmd.Flags().StringVar(&assetOpts.ipMacJSON, "ip-mac", "", "asset IP/MAC bindings as JSON array, for example '[{\"ip\":[\"192.0.2.10\"],\"mac\":[\"00:11:22:33:44:55\"]}]'")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newAssetUpdateCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetUpdateOptions
	cmd := &cobra.Command{
		Use:   "update",
		Short: "编辑资产",
		Long: "编辑资产，用于更新单个资产基础信息。该命令是高影响写操作：预览阶段会读取当前资产详情并返回 before/after 摘要；必须使用 --confirm CONFIRM_ASSET_UPDATE 才会调用后端写入接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer asset update --id 9 --name core-db-new --ip 192.0.2.11 --preview\n" +
			"  chaitin-cli tanswer asset update --id 9 --name core-db-new --ip 192.0.2.11 --confirm CONFIRM_ASSET_UPDATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetUpdateRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "编辑资产", "chaitin-cli tanswer asset update", "INVALID_ASSET_UPDATE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(assetOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(assetOpts.confirm, assetUpdateConfirmToken); err != nil {
					return writeAssetError(cmd, "编辑资产", "chaitin-cli tanswer asset update", "ASSET_UPDATE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchAssetDetail(cmd, client, req["id"].(uint))
			if err != nil {
				return writeAssetError(cmd, "编辑资产", "chaitin-cli tanswer asset update", "ASSET_UPDATE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{
				"before": before,
				"after":  req,
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "编辑资产预览",
				Command:       "chaitin-cli tanswer asset update",
				OperationType: "asset_update",
				RiskLevel:     "write_high",
				Target: map[string]any{
					"id": req["id"],
				},
				ChangeSummary: changeSummary,
				Impact: map[string]any{
					"asset_count": 1,
					"group_id":    req["group_id"],
				},
				RiskWarnings: []string{
					"将更新资产配置并触发资产配置版本更新。",
					"如果 IP 与其他资产冲突，后端会拒绝写入。",
				},
				ConfirmToken: assetUpdateConfirmToken,
			})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{
					Success: true,
					Task:    "编辑资产预览",
					Command: "chaitin-cli tanswer asset update",
					Query:   req,
					Data:    preview,
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.UpdateAsset", req, &struct{}{}); err != nil {
				return writeAssetError(cmd, "编辑资产", "chaitin-cli tanswer asset update", "ASSET_UPDATE_FAILED", err.Error(), true)
			}
			object := map[string]any{
				"id":         req["id"],
				"name":       req["name"],
				"ip":         req["ip"],
				"group_id":   req["group_id"],
				"importance": req["importance"],
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "asset_update",
				Object:        object,
				Action:        "update",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "编辑资产",
				Command: "chaitin-cli tanswer asset update",
				Query:   req,
				Data:    data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.id, "id", "", "asset id, required")
	cmd.Flags().StringVar(&assetOpts.name, "name", "", "asset name, required")
	cmd.Flags().StringVar(&assetOpts.ip, "ip", "", "asset IP list, comma separated, required")
	cmd.Flags().StringVar(&assetOpts.contact, "contact", "", "asset owner/contact")
	cmd.Flags().StringVar(&assetOpts.importance, "importance", "normal", "asset importance: important, normal, 重点, 普通, 1, or 2")
	cmd.Flags().StringVar(&assetOpts.remark, "remark", "", "asset remark")
	cmd.Flags().StringVar(&assetOpts.assetType, "asset-type", "", "asset type")
	cmd.Flags().StringVar(&assetOpts.location, "location", "", "asset location")
	cmd.Flags().StringVar(&assetOpts.tagID, "tag-id", "", "asset tag IDs, comma separated")
	cmd.Flags().UintVar(&assetOpts.groupID, "group-id", 1, "asset group ID")
	cmd.Flags().StringVar(&assetOpts.ipMacJSON, "ip-mac", "", "asset IP/MAC bindings as JSON array")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newAssetDeleteCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetDeleteOptions
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "删除资产",
		Long: "删除资产，用于移除一个或多个已有资产配置。该命令是高影响写操作：预览阶段会读取待删除资产详情并返回 before 摘要；必须使用 --confirm CONFIRM_ASSET_DELETE 才会调用后端写入接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer asset delete --id-list 9 --preview\n" +
			"  chaitin-cli tanswer asset delete --id-list 9,10 --confirm CONFIRM_ASSET_DELETE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetDeleteRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "删除资产", "chaitin-cli tanswer asset delete", "INVALID_ASSET_DELETE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(assetOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(assetOpts.confirm, assetDeleteConfirmToken); err != nil {
					return writeAssetError(cmd, "删除资产", "chaitin-cli tanswer asset delete", "ASSET_DELETE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			ids := req["ids"].([]uint)
			before, err := fetchAssetDetails(cmd, client, ids)
			if err != nil {
				return writeAssetError(cmd, "删除资产", "chaitin-cli tanswer asset delete", "ASSET_DELETE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{
				"before": before,
				"after":  nil,
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "删除资产预览",
				Command:       "chaitin-cli tanswer asset delete",
				OperationType: "asset_delete",
				RiskLevel:     "write_high",
				Target: map[string]any{
					"ids": ids,
				},
				ChangeSummary: changeSummary,
				Impact: map[string]any{
					"asset_count": len(ids),
				},
				RiskWarnings: []string{
					"将删除资产配置并触发资产配置版本更新。",
					"删除资产可能影响后续资产筛选、告警归属和运营查询结果。",
				},
				ConfirmToken: assetDeleteConfirmToken,
			})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{
					Success: true,
					Task:    "删除资产预览",
					Command: "chaitin-cli tanswer asset delete",
					Query:   req,
					Data:    preview,
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.DeleteAsset", req, &struct{}{}); err != nil {
				return writeAssetError(cmd, "删除资产", "chaitin-cli tanswer asset delete", "ASSET_DELETE_FAILED", err.Error(), true)
			}
			object := map[string]any{
				"ids": ids,
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "asset_delete",
				Object:        object,
				Action:        "delete",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "删除资产",
				Command: "chaitin-cli tanswer asset delete",
				Query:   req,
				Data:    data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.idList, "id-list", "", "comma-separated asset IDs to delete, required")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newAssetBatchMaintainCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetBatchMaintainOptions
	cmd := &cobra.Command{
		Use:   "batch-maintain",
		Short: "批量维护资产",
		Long: "批量维护资产，用于批量更新已有资产的资产组、负责人、地理位置或备注。该命令是高影响写操作：预览阶段会读取待维护资产详情并返回 before/after 摘要；必须使用 --confirm CONFIRM_ASSET_BATCH_MAINTAIN 才会调用后端写入接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer asset batch-maintain --id-list 9,10 --contact secops --preview\n" +
			"  chaitin-cli tanswer asset batch-maintain --id-list 9,10 --group-id 2 --confirm CONFIRM_ASSET_BATCH_MAINTAIN",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetBatchMaintainRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "批量维护资产", "chaitin-cli tanswer asset batch-maintain", "INVALID_ASSET_BATCH_MAINTAIN_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(assetOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(assetOpts.confirm, assetBatchMaintainConfirmToken); err != nil {
					return writeAssetError(cmd, "批量维护资产", "chaitin-cli tanswer asset batch-maintain", "ASSET_BATCH_MAINTAIN_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			ids := req["ids"].([]uint)
			before, err := fetchAssetDetails(cmd, client, ids)
			if err != nil {
				return writeAssetError(cmd, "批量维护资产", "chaitin-cli tanswer asset batch-maintain", "ASSET_BATCH_MAINTAIN_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{
				"before": before,
				"after":  assetBatchMaintainChanges(req),
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "批量维护资产预览",
				Command:       "chaitin-cli tanswer asset batch-maintain",
				OperationType: "asset_batch_maintain",
				RiskLevel:     "write_high",
				Target: map[string]any{
					"ids": ids,
				},
				ChangeSummary: changeSummary,
				Impact: map[string]any{
					"asset_count": len(ids),
				},
				RiskWarnings: []string{
					"将批量更新资产配置并触发资产配置版本更新。",
					"批量修改资产组可能影响资产树、告警归属和后续筛选结果。",
				},
				ConfirmToken: assetBatchMaintainConfirmToken,
			})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{
					Success: true,
					Task:    "批量维护资产预览",
					Command: "chaitin-cli tanswer asset batch-maintain",
					Query:   req,
					Data:    preview,
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.UpdateAssetBatch", req, &struct{}{}); err != nil {
				return writeAssetError(cmd, "批量维护资产", "chaitin-cli tanswer asset batch-maintain", "ASSET_BATCH_MAINTAIN_FAILED", err.Error(), true)
			}
			object := map[string]any{
				"ids": ids,
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "asset_batch_maintain",
				Object:        object,
				Action:        "batch_maintain",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "批量维护资产",
				Command: "chaitin-cli tanswer asset batch-maintain",
				Query:   req,
				Data:    data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.idList, "id-list", "", "comma-separated asset IDs to maintain, required")
	cmd.Flags().StringVar(&assetOpts.contact, "contact", "", "asset owner/contact to set")
	cmd.Flags().StringVar(&assetOpts.remark, "remark", "", "asset remark to set")
	cmd.Flags().StringVar(&assetOpts.location, "location", "", "asset location to set")
	cmd.Flags().UintVar(&assetOpts.groupID, "group-id", 0, "asset group ID to move assets into")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newAssetBatchTagCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetBatchTagOptions
	cmd := &cobra.Command{
		Use:   "batch-tag",
		Short: "批量维护资产标签",
		Long: "批量维护资产标签，用于给一个或多个已有资产批量设置资产标签。该命令是高影响写操作：预览阶段会读取待维护资产详情并返回 before/after 摘要；必须使用 --confirm CONFIRM_ASSET_BATCH_TAG 才会调用后端写入接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer asset batch-tag --id-list 9,10 --tag-id 3,7 --preview\n" +
			"  chaitin-cli tanswer asset batch-tag --id-list 9,10 --tag-id 3,7 --confirm CONFIRM_ASSET_BATCH_TAG",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetBatchTagRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "批量维护资产标签", "chaitin-cli tanswer asset batch-tag", "INVALID_ASSET_BATCH_TAG_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(assetOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(assetOpts.confirm, assetBatchTagConfirmToken); err != nil {
					return writeAssetError(cmd, "批量维护资产标签", "chaitin-cli tanswer asset batch-tag", "ASSET_BATCH_TAG_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			ids := req["ids"].([]uint)
			before, err := fetchAssetDetails(cmd, client, ids)
			if err != nil {
				return writeAssetError(cmd, "批量维护资产标签", "chaitin-cli tanswer asset batch-tag", "ASSET_BATCH_TAG_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{
				"before": before,
				"after":  map[string]any{"tag_ids": req["tag_ids"]},
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "批量维护资产标签预览",
				Command:       "chaitin-cli tanswer asset batch-tag",
				OperationType: "asset_batch_tag",
				RiskLevel:     "write_high",
				Target:        map[string]any{"ids": ids},
				ChangeSummary: changeSummary,
				Impact: map[string]any{
					"asset_count": len(ids),
					"tag_ids":     req["tag_ids"],
				},
				RiskWarnings: []string{
					"将批量更新资产标签配置。",
					"标签变化可能影响后续资产筛选、运营分组和查询结果。",
				},
				ConfirmToken: assetBatchTagConfirmToken,
			})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{
					Success: true,
					Task:    "批量维护资产标签预览",
					Command: "chaitin-cli tanswer asset batch-tag",
					Query:   req,
					Data:    preview,
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.UpdateAssetTagBatch", req, &struct{}{}); err != nil {
				return writeAssetError(cmd, "批量维护资产标签", "chaitin-cli tanswer asset batch-tag", "ASSET_BATCH_TAG_FAILED", err.Error(), true)
			}
			object := map[string]any{
				"ids":     ids,
				"tag_ids": req["tag_ids"],
			}
			data := BuildWriteExecutionResult(WriteExecutionSpec{
				OperationType: "asset_batch_tag",
				Object:        object,
				Action:        "batch_tag",
				Environment:   cfg.BaseURL,
				Actor:         "open_api_token",
				BeforeAfter:   changeSummary,
				Result:        "success",
			})
			raw, err := RenderJSON(SuccessEnvelope{
				Success: true,
				Task:    "批量维护资产标签",
				Command: "chaitin-cli tanswer asset batch-tag",
				Query:   req,
				Data:    data,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.idList, "id-list", "", "comma-separated asset IDs to maintain, required")
	cmd.Flags().StringVar(&assetOpts.tagID, "tag-id", "", "asset tag IDs to set, comma separated, required")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newAssetGroupCreateCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetGroupCreateOptions
	cmd := &cobra.Command{
		Use:   "group-create",
		Short: "创建资产组",
		Long: "创建资产组，用于在指定父资产组下新增一个资产组。该命令是高影响写操作：预览阶段会读取父资产组摘要；必须使用 --confirm CONFIRM_ASSET_GROUP_CREATE 才会调用后端写入接口。\n\n" +
			"输出预览：requires_confirmation、confirmed、operation_type、target、change_summary、impact、risk_warnings、confirmation_token。\n" +
			"执行输出：confirmed、result、object、audit。",
		Example: "  chaitin-cli tanswer asset group-create --name 核心区 --parent-id 2 --preview\n" +
			"  chaitin-cli tanswer asset group-create --name 核心区 --parent-id 2 --confirm CONFIRM_ASSET_GROUP_CREATE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetGroupCreateRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "创建资产组", "chaitin-cli tanswer asset group-create", "INVALID_ASSET_GROUP_CREATE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(assetOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(assetOpts.confirm, assetGroupCreateConfirmToken); err != nil {
					return writeAssetError(cmd, "创建资产组", "chaitin-cli tanswer asset group-create", "ASSET_GROUP_CREATE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			parent, err := fetchAssetGroupDetail(cmd, client, req["pid"].(uint))
			if err != nil {
				return writeAssetError(cmd, "创建资产组", "chaitin-cli tanswer asset group-create", "ASSET_GROUP_CREATE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{"before": parent, "after": req}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "创建资产组预览",
				Command:       "chaitin-cli tanswer asset group-create",
				OperationType: "asset_group_create",
				RiskLevel:     "write_high",
				Target:        map[string]any{"parent_id": req["pid"], "name": req["name"]},
				ChangeSummary: changeSummary,
				Impact:        map[string]any{"group_count": 1, "parent_id": req["pid"]},
				RiskWarnings:  []string{"将新增资产组并改变资产组树结构。"},
				ConfirmToken:  assetGroupCreateConfirmToken,
			})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "创建资产组预览", Command: "chaitin-cli tanswer asset group-create", Query: req, Data: preview})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			var result struct {
				ID uint `json:"id"`
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.CreateGroup", req, &result); err != nil {
				return writeAssetError(cmd, "创建资产组", "chaitin-cli tanswer asset group-create", "ASSET_GROUP_CREATE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": result.ID, "name": req["name"], "parent_id": req["pid"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "asset_group_create", Object: object, Action: "group_create", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: changeSummary, Result: "success"})
			raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "创建资产组", Command: "chaitin-cli tanswer asset group-create", Query: req, Data: data})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.name, "name", "", "asset group name, required")
	cmd.Flags().UintVar(&assetOpts.parentID, "parent-id", 1, "parent asset group ID")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newAssetGroupRenameCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetGroupRenameOptions
	cmd := &cobra.Command{
		Use:   "group-rename",
		Short: "重命名资产组",
		Long:  "重命名资产组，用于修改一个已有资产组名称。该命令是高影响写操作：预览阶段会读取当前资产组摘要并返回 before/after；必须使用 --confirm CONFIRM_ASSET_GROUP_RENAME 才会调用后端写入接口。",
		Example: "  chaitin-cli tanswer asset group-rename --id 3 --name 核心区 --preview\n" +
			"  chaitin-cli tanswer asset group-rename --id 3 --name 核心区 --confirm CONFIRM_ASSET_GROUP_RENAME",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetGroupRenameRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "重命名资产组", "chaitin-cli tanswer asset group-rename", "INVALID_ASSET_GROUP_RENAME_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(assetOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(assetOpts.confirm, assetGroupRenameConfirmToken); err != nil {
					return writeAssetError(cmd, "重命名资产组", "chaitin-cli tanswer asset group-rename", "ASSET_GROUP_RENAME_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			before, err := fetchAssetGroupDetail(cmd, client, req["id"].(uint))
			if err != nil {
				return writeAssetError(cmd, "重命名资产组", "chaitin-cli tanswer asset group-rename", "ASSET_GROUP_RENAME_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{"before": before, "after": req}
			preview := BuildWritePreview(WriteOperationSpec{Task: "重命名资产组预览", Command: "chaitin-cli tanswer asset group-rename", OperationType: "asset_group_rename", RiskLevel: "write_high", Target: map[string]any{"id": req["id"]}, ChangeSummary: changeSummary, Impact: map[string]any{"group_count": 1}, RiskWarnings: []string{"将修改资产组名称并影响后续资产组查询显示。"}, ConfirmToken: assetGroupRenameConfirmToken})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "重命名资产组预览", Command: "chaitin-cli tanswer asset group-rename", Query: req, Data: preview})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.UpdateGroup", req, &struct{}{}); err != nil {
				return writeAssetError(cmd, "重命名资产组", "chaitin-cli tanswer asset group-rename", "ASSET_GROUP_RENAME_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": req["id"], "name": req["name"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "asset_group_rename", Object: object, Action: "group_rename", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: changeSummary, Result: "success"})
			raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "重命名资产组", Command: "chaitin-cli tanswer asset group-rename", Query: req, Data: data})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.id, "id", "", "asset group id, required")
	cmd.Flags().StringVar(&assetOpts.name, "name", "", "asset group name, required")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newAssetGroupDeleteCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetGroupDeleteOptions
	cmd := &cobra.Command{
		Use:   "group-delete",
		Short: "删除资产组",
		Long:  "删除资产组，用于删除一个或多个非根资产组。该命令是高影响写操作：预览阶段会读取待删除资产组摘要；必须使用 --confirm CONFIRM_ASSET_GROUP_DELETE 才会调用后端写入接口。",
		Example: "  chaitin-cli tanswer asset group-delete --id-list 3 --preview\n" +
			"  chaitin-cli tanswer asset group-delete --id-list 3,4 --confirm CONFIRM_ASSET_GROUP_DELETE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetGroupDeleteRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "删除资产组", "chaitin-cli tanswer asset group-delete", "INVALID_ASSET_GROUP_DELETE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(assetOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(assetOpts.confirm, assetGroupDeleteConfirmToken); err != nil {
					return writeAssetError(cmd, "删除资产组", "chaitin-cli tanswer asset group-delete", "ASSET_GROUP_DELETE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			ids := req["ids"].([]uint)
			before, err := fetchAssetGroupDetails(cmd, client, ids)
			if err != nil {
				return writeAssetError(cmd, "删除资产组", "chaitin-cli tanswer asset group-delete", "ASSET_GROUP_DELETE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{"before": before, "after": nil}
			preview := BuildWritePreview(WriteOperationSpec{Task: "删除资产组预览", Command: "chaitin-cli tanswer asset group-delete", OperationType: "asset_group_delete", RiskLevel: "write_high", Target: map[string]any{"ids": ids}, ChangeSummary: changeSummary, Impact: map[string]any{"group_count": len(ids)}, RiskWarnings: []string{"将删除资产组并改变资产组树结构。", "后端会按产品规则处理组内资产或子组。"}, ConfirmToken: assetGroupDeleteConfirmToken})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "删除资产组预览", Command: "chaitin-cli tanswer asset group-delete", Query: req, Data: preview})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.DeleteGroup", req, &struct{}{}); err != nil {
				return writeAssetError(cmd, "删除资产组", "chaitin-cli tanswer asset group-delete", "ASSET_GROUP_DELETE_FAILED", err.Error(), true)
			}
			object := map[string]any{"ids": ids}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "asset_group_delete", Object: object, Action: "group_delete", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: changeSummary, Result: "success"})
			raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "删除资产组", Command: "chaitin-cli tanswer asset group-delete", Query: req, Data: data})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.idList, "id-list", "", "comma-separated asset group IDs to delete, required")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func newAssetTreeMoveCommand(opts *RootOptions) *cobra.Command {
	var assetOpts assetTreeMoveOptions
	cmd := &cobra.Command{
		Use:   "tree-move",
		Short: "调整资产树层级",
		Long: "调整资产树层级，用于移动资产或资产组在资产树中的位置。该命令是高影响写操作：预览阶段会读取源节点和前置/目标节点摘要；必须使用 --confirm CONFIRM_ASSET_TREE_MOVE 才会调用后端写入接口。\n\n" +
			"节点类型：group/1 表示资产组，asset/2 表示资产。--top-layer 对应产品拖拽参数 top_layer。",
		Example: "  chaitin-cli tanswer asset tree-move --id 9 --type asset --prev-id 3 --prev-type group --top-layer --preview\n" +
			"  chaitin-cli tanswer asset tree-move --id 9 --type asset --prev-id 3 --prev-type group --top-layer --confirm CONFIRM_ASSET_TREE_MOVE",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := buildAssetTreeMoveRequest(assetOpts)
			if err != nil {
				return writeAssetError(cmd, "调整资产树层级", "chaitin-cli tanswer asset tree-move", "INVALID_ASSET_TREE_MOVE_REQUEST", err.Error(), false)
			}
			if strings.TrimSpace(assetOpts.confirm) != "" {
				if err := ValidateWriteConfirmation(assetOpts.confirm, assetTreeMoveConfirmToken); err != nil {
					return writeAssetError(cmd, "调整资产树层级", "chaitin-cli tanswer asset tree-move", "ASSET_TREE_MOVE_CONFIRMATION_REQUIRED", err.Error(), false)
				}
			}
			cfg, err := LoadConfig(ConfigOptions{Address: opts.Address, Token: opts.Token, Timeout: opts.Timeout, InsecureSkipVerify: opts.InsecureSkipVerify, InsecureSkipVerifySet: opts.InsecureSkipVerifySet})
			if err != nil {
				return err
			}
			client := NewClient(cfg)
			source, err := fetchAssetTreeNodeDetail(cmd, client, req["id"].(uint), req["type"].(int))
			if err != nil {
				return writeAssetError(cmd, "调整资产树层级", "chaitin-cli tanswer asset tree-move", "ASSET_TREE_MOVE_PREVIEW_FAILED", err.Error(), true)
			}
			prev, err := fetchAssetTreeNodeDetail(cmd, client, req["prev_id"].(uint), req["prev_type"].(int))
			if err != nil {
				return writeAssetError(cmd, "调整资产树层级", "chaitin-cli tanswer asset tree-move", "ASSET_TREE_MOVE_PREVIEW_FAILED", err.Error(), true)
			}
			changeSummary := map[string]any{
				"before": map[string]any{"source": source, "prev": prev},
				"after":  req,
			}
			preview := BuildWritePreview(WriteOperationSpec{
				Task:          "调整资产树层级预览",
				Command:       "chaitin-cli tanswer asset tree-move",
				OperationType: "asset_tree_move",
				RiskLevel:     "write_high",
				Target:        map[string]any{"id": req["id"], "type": req["type"]},
				ChangeSummary: changeSummary,
				Impact:        map[string]any{"moved_count": 1, "top_layer": req["top_layer"]},
				RiskWarnings:  []string{"将调整资产树层级，可能影响资产组层级、资产筛选和运营查询结果。"},
				ConfirmToken:  assetTreeMoveConfirmToken,
			})
			if assetOpts.preview || strings.TrimSpace(assetOpts.confirm) == "" {
				raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "调整资产树层级预览", Command: "chaitin-cli tanswer asset tree-move", Query: req, Data: preview})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}
			if err := client.CallRPC(cmd.Context(), "AssetService.UpdateAssetTree", req, &struct{}{}); err != nil {
				return writeAssetError(cmd, "调整资产树层级", "chaitin-cli tanswer asset tree-move", "ASSET_TREE_MOVE_FAILED", err.Error(), true)
			}
			object := map[string]any{"id": req["id"], "type": req["type"], "prev_id": req["prev_id"], "prev_type": req["prev_type"], "top_layer": req["top_layer"]}
			data := BuildWriteExecutionResult(WriteExecutionSpec{OperationType: "asset_tree_move", Object: object, Action: "tree_move", Environment: cfg.BaseURL, Actor: "open_api_token", BeforeAfter: changeSummary, Result: "success"})
			raw, err := RenderJSON(SuccessEnvelope{Success: true, Task: "调整资产树层级", Command: "chaitin-cli tanswer asset tree-move", Query: req, Data: data})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	cmd.Flags().StringVar(&assetOpts.id, "id", "", "source node id, required")
	cmd.Flags().StringVar(&assetOpts.nodeType, "type", "", "source node type: group, asset, 1, or 2, required")
	cmd.Flags().StringVar(&assetOpts.prevID, "prev-id", "", "previous/target node id, required")
	cmd.Flags().StringVar(&assetOpts.prevType, "prev-type", "", "previous/target node type: group, asset, 1, or 2, required")
	cmd.Flags().BoolVar(&assetOpts.topLayer, "top-layer", false, "backend top_layer flag from product tree drag")
	cmd.Flags().BoolVar(&assetOpts.preview, "preview", false, "return write preview without executing")
	cmd.Flags().StringVar(&assetOpts.confirm, "confirm", "", "exact confirmation token required to execute")
	return cmd
}

func addAssetListFlags(cmd *cobra.Command, opts *assetListOptions) {
	cmd.Flags().IntVar(&opts.page, "page", 1, "page number, starts from 1")
	cmd.Flags().IntVar(&opts.pageSize, "page-size", 10, "page size, 1-100")
	cmd.Flags().UintVar(&opts.id, "id", 0, "asset id")
	cmd.Flags().StringVar(&opts.name, "name", "", "asset name fuzzy filter")
	cmd.Flags().StringVar(&opts.ip, "ip", "", "asset IP, CIDR, or range fuzzy filter")
	cmd.Flags().StringVar(&opts.mac, "mac", "", "asset MAC fuzzy filter")
	cmd.Flags().StringVar(&opts.assetType, "asset-type", "", "asset type fuzzy filter")
	cmd.Flags().StringVar(&opts.importance, "importance", "", "asset importance: important, normal, 重点, 普通, 1, 2")
	cmd.Flags().StringVar(&opts.tagID, "tag-id", "", "asset tag id filter, comma separated")
	cmd.Flags().UintVar(&opts.groupID, "group-id", 0, "asset group id filter")
}

func buildAssetGroupTreeRequest(opts assetGroupTreeOptions) map[string]any {
	return map[string]any{
		"id":         opts.id,
		"depth":      opts.depth,
		"with_asset": opts.withAsset,
	}
}

func buildAssetDownloadTemplateRequest(opts assetDownloadTemplateOptions) map[string]any {
	return map[string]any{
		"with_data":    false,
		"with_example": opts.withExample,
		"id_list":      []int64{},
	}
}

func buildAssetImportRequest(opts assetImportOptions) (map[string]any, error) {
	path := strings.TrimSpace(opts.filePath)
	if path == "" {
		return nil, fmt.Errorf("missing import file: set --file")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("import file must not be a directory")
	}
	return map[string]any{
		"file_path":  path,
		"file_name":  filepath.Base(path),
		"size_bytes": info.Size(),
		"method":     assetImportMethod,
	}, nil
}

func buildAssetExportRequest(opts assetExportOptions) (map[string]any, error) {
	ids, err := parseInt64CSV(opts.idList)
	if err != nil {
		return nil, fmt.Errorf("invalid id-list: %w", err)
	}
	return map[string]any{
		"with_data":    true,
		"with_example": false,
		"id_list":      ids,
	}, nil
}

func buildAssetCreateRequest(opts assetCreateOptions) (map[string]any, error) {
	name := strings.TrimSpace(opts.name)
	if name == "" {
		return nil, fmt.Errorf("missing asset name: set --name")
	}
	if strings.Contains(name, " ") {
		return nil, fmt.Errorf("asset name must not contain spaces")
	}
	ip := strings.TrimSpace(opts.ip)
	if ip == "" {
		return nil, fmt.Errorf("missing asset IP: set --ip")
	}
	importance, err := assetImportanceValue(opts.importance)
	if err != nil {
		return nil, err
	}
	tagIDs, err := parseUintCSV(opts.tagID)
	if err != nil {
		return nil, fmt.Errorf("invalid tag-id: %w", err)
	}
	ipMac, err := parseAssetCreateIPMac(opts.ipMacJSON)
	if err != nil {
		return nil, err
	}
	groupID := opts.groupID
	if groupID == 0 {
		groupID = 1
	}
	return map[string]any{
		"name":        name,
		"ip":          strings.ReplaceAll(ip, " ", ""),
		"contact":     strings.TrimSpace(opts.contact),
		"importance":  importance,
		"remark":      strings.TrimSpace(opts.remark),
		"asset_type":  strings.TrimSpace(opts.assetType),
		"location":    strings.TrimSpace(opts.location),
		"tag_id_list": tagIDs,
		"group_id":    groupID,
		"ip_mac":      ipMac,
	}, nil
}

func buildAssetUpdateRequest(opts assetUpdateOptions) (map[string]any, error) {
	id, err := parseRequiredAssetID(opts.id)
	if err != nil {
		return nil, err
	}
	req, err := buildAssetCreateRequest(assetCreateOptions{
		name:       opts.name,
		ip:         opts.ip,
		contact:    opts.contact,
		importance: opts.importance,
		remark:     opts.remark,
		assetType:  opts.assetType,
		location:   opts.location,
		tagID:      opts.tagID,
		groupID:    opts.groupID,
		ipMacJSON:  opts.ipMacJSON,
	})
	if err != nil {
		return nil, err
	}
	req["id"] = id
	return req, nil
}

func buildAssetDeleteRequest(opts assetDeleteOptions) (map[string]any, error) {
	ids, err := parseUintCSV(opts.idList)
	if err != nil {
		return nil, fmt.Errorf("invalid id-list: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("missing asset ids: set --id-list")
	}
	for _, id := range ids {
		if id == 0 {
			return nil, fmt.Errorf("asset id must be greater than 0")
		}
	}
	return map[string]any{"ids": ids}, nil
}

func buildAssetBatchMaintainRequest(opts assetBatchMaintainOptions) (map[string]any, error) {
	ids, err := parseUintCSV(opts.idList)
	if err != nil {
		return nil, fmt.Errorf("invalid id-list: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("missing asset ids: set --id-list")
	}
	for _, id := range ids {
		if id == 0 {
			return nil, fmt.Errorf("asset id must be greater than 0")
		}
	}
	req := map[string]any{"ids": ids}
	if value := strings.TrimSpace(opts.contact); value != "" {
		req["contact"] = value
	}
	if value := strings.TrimSpace(opts.remark); value != "" {
		req["remark"] = value
	}
	if value := strings.TrimSpace(opts.location); value != "" {
		req["location"] = value
	}
	if opts.groupID != 0 {
		req["group_id"] = opts.groupID
	}
	if len(req) == 1 {
		return nil, fmt.Errorf("missing fields to update: set one of --contact, --remark, --location, or --group-id")
	}
	return req, nil
}

func buildAssetBatchTagRequest(opts assetBatchTagOptions) (map[string]any, error) {
	ids, err := parseUintCSV(opts.idList)
	if err != nil {
		return nil, fmt.Errorf("invalid id-list: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("missing asset ids: set --id-list")
	}
	for _, id := range ids {
		if id == 0 {
			return nil, fmt.Errorf("asset id must be greater than 0")
		}
	}
	tagIDs, err := parseUintCSV(opts.tagID)
	if err != nil {
		return nil, fmt.Errorf("invalid tag-id: %w", err)
	}
	if len(tagIDs) == 0 {
		return nil, fmt.Errorf("missing tag ids: set --tag-id")
	}
	for _, id := range tagIDs {
		if id == 0 {
			return nil, fmt.Errorf("tag id must be greater than 0")
		}
	}
	return map[string]any{
		"ids":     ids,
		"tag_ids": tagIDs,
	}, nil
}

func buildAssetGroupCreateRequest(opts assetGroupCreateOptions) (map[string]any, error) {
	name := strings.TrimSpace(opts.name)
	if name == "" {
		return nil, fmt.Errorf("missing asset group name: set --name")
	}
	parentID := opts.parentID
	if parentID == 0 {
		parentID = 1
	}
	return map[string]any{"name": name, "pid": parentID}, nil
}

func buildAssetGroupRenameRequest(opts assetGroupRenameOptions) (map[string]any, error) {
	id, err := parseRequiredAssetID(opts.id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(opts.name)
	if name == "" {
		return nil, fmt.Errorf("missing asset group name: set --name")
	}
	return map[string]any{"id": id, "name": name}, nil
}

func buildAssetGroupDeleteRequest(opts assetGroupDeleteOptions) (map[string]any, error) {
	ids, err := parseUintCSV(opts.idList)
	if err != nil {
		return nil, fmt.Errorf("invalid id-list: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("missing asset group ids: set --id-list")
	}
	for _, id := range ids {
		if id == 0 {
			return nil, fmt.Errorf("asset group id must be greater than 0")
		}
		if id == 1 {
			return nil, fmt.Errorf("root asset group cannot be deleted")
		}
	}
	return map[string]any{"ids": ids}, nil
}

func buildAssetTreeMoveRequest(opts assetTreeMoveOptions) (map[string]any, error) {
	id, err := parseRequiredAssetID(opts.id)
	if err != nil {
		return nil, err
	}
	nodeType, err := parseAssetTreeNodeType(opts.nodeType)
	if err != nil {
		return nil, fmt.Errorf("invalid type: %w", err)
	}
	prevID, err := parseRequiredAssetID(opts.prevID)
	if err != nil {
		return nil, fmt.Errorf("invalid prev-id: %w", err)
	}
	prevType, err := parseAssetTreeNodeType(opts.prevType)
	if err != nil {
		return nil, fmt.Errorf("invalid prev-type: %w", err)
	}
	return map[string]any{
		"id":        id,
		"type":      nodeType,
		"prev_id":   prevID,
		"prev_type": prevType,
		"top_layer": opts.topLayer,
	}, nil
}

func buildAssetListRequest(opts assetListOptions) (map[string]any, error) {
	offset := int64((opts.page - 1) * opts.pageSize)
	count := int64(opts.pageSize)
	req := map[string]any{
		"offset": offset,
		"count":  count,
	}
	for key, value := range assetListFilters(opts) {
		req[key] = value
	}
	if strings.TrimSpace(opts.importance) != "" {
		importance, err := assetImportanceValue(opts.importance)
		if err != nil {
			return nil, err
		}
		req["importance"] = importance
	}
	if strings.TrimSpace(opts.tagID) != "" {
		tagIDs, err := parseUintCSV(opts.tagID)
		if err != nil {
			return nil, fmt.Errorf("invalid tag-id: %w", err)
		}
		req["tag_id"] = tagIDs
	}
	return req, nil
}

func downloadAssetFile(cmd *cobra.Command, cfg Config, task string, command string, query map[string]any, outputPath string, defaultName string, exportScope string) error {
	client := NewClient(cfg)
	file, err := client.Download(cmd.Context(), assetDownloadMethod, query)
	if err != nil {
		return writeAssetError(cmd, task, command, "ASSET_FILE_DOWNLOAD_FAILED", err.Error(), true)
	}
	if file.StatusCode < 200 || file.StatusCode >= 300 {
		return writeAssetError(cmd, task, command, "ASSET_FILE_DOWNLOAD_FAILED", fmt.Sprintf("download returned HTTP %d: %s", file.StatusCode, strings.TrimSpace(string(file.Body))), true)
	}
	target, fileName := resolveDownloadOutputPath(outputPath, file.FileName, defaultName)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return writeAssetError(cmd, task, command, "ASSET_FILE_WRITE_FAILED", err.Error(), false)
	}
	if err := os.WriteFile(target, file.Body, 0644); err != nil {
		return writeAssetError(cmd, task, command, "ASSET_FILE_WRITE_FAILED", err.Error(), false)
	}
	data := map[string]any{
		"file_name":      fileName,
		"file_path":      target,
		"size_bytes":     len(file.Body),
		"status_code":    file.StatusCode,
		"method":         assetDownloadMethod,
		"download_query": query,
	}
	if exportScope != "" {
		data["export_scope"] = exportScope
	}
	raw, err := RenderJSON(SuccessEnvelope{
		Success: true,
		Task:    task,
		Command: command,
		Query:   query,
		Data:    data,
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

func fetchAssetDetail(cmd *cobra.Command, client *Client, id uint) (map[string]any, error) {
	var result assetDetailRPCResult
	if err := client.CallRPC(cmd.Context(), "AssetService.GetAssetInfo", map[string]any{"id": id}, &result); err != nil {
		return nil, err
	}
	if assetDetailIsMissing(result.Data) {
		return nil, fmt.Errorf("asset %d was not found", id)
	}
	return result.Data, nil
}

func assetDetailIsMissing(detail map[string]any) bool {
	if len(detail) == 0 {
		return true
	}
	value, ok := detail["id"]
	if !ok {
		return false
	}
	switch id := value.(type) {
	case float64:
		return id == 0
	case float32:
		return id == 0
	case int:
		return id == 0
	case int64:
		return id == 0
	case uint:
		return id == 0
	case uint64:
		return id == 0
	case string:
		return strings.TrimSpace(id) == "0"
	default:
		return false
	}
}

func fetchAssetDetails(cmd *cobra.Command, client *Client, ids []uint) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		detail, err := fetchAssetDetail(cmd, client, id)
		if err != nil {
			return nil, err
		}
		out = append(out, detail)
	}
	return out, nil
}

func fetchAssetGroupDetail(cmd *cobra.Command, client *Client, id uint) (map[string]any, error) {
	req := map[string]any{"id": id, "depth": 1, "with_asset": true}
	var result assetGroupTreeRPCResult
	if err := client.CallRPC(cmd.Context(), "AssetService.SearchGroups", req, &result); err != nil {
		return nil, err
	}
	groups := summarizeAssetGroupTree(result.Data, nil)
	if len(groups) == 0 {
		return map[string]any{"id": id, "exists": false}, nil
	}
	return groups[0], nil
}

func fetchAssetGroupDetails(cmd *cobra.Command, client *Client, ids []uint) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		detail, err := fetchAssetGroupDetail(cmd, client, id)
		if err != nil {
			return nil, err
		}
		out = append(out, detail)
	}
	return out, nil
}

func fetchAssetTreeNodeDetail(cmd *cobra.Command, client *Client, id uint, nodeType int) (map[string]any, error) {
	switch nodeType {
	case assetTreeNodeTypeAsset:
		return fetchAssetDetail(cmd, client, id)
	case assetTreeNodeTypeGroup:
		return fetchAssetGroupDetail(cmd, client, id)
	default:
		return nil, fmt.Errorf("unsupported node type %d", nodeType)
	}
}

func assetBatchMaintainChanges(req map[string]any) map[string]any {
	changes := map[string]any{}
	for _, key := range []string{"contact", "remark", "location", "group_id"} {
		if value, ok := req[key]; ok {
			changes[key] = value
		}
	}
	return changes
}

func resolveDownloadOutputPath(outputPath string, downloadedName string, defaultName string) (string, string) {
	fileName := firstNonEmpty(downloadedName, defaultName)
	target := strings.TrimSpace(outputPath)
	if target == "" {
		target = fileName
	}
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		target = filepath.Join(target, fileName)
	}
	return target, filepath.Base(target)
}

func summarizeAssetGroupTree(nodes []assetGroupTreeNode, parentPath []string) []map[string]any {
	out := make([]map[string]any, 0, len(nodes))
	for _, node := range nodes {
		pathParts := append(append([]string{}, parentPath...), node.Name)
		summary := map[string]any{
			"id":         node.ID,
			"name":       node.Name,
			"type":       node.Type,
			"type_label": assetGroupTreeTypeLabel(node.Type),
			"count":      node.Count,
			"path":       strings.Join(pathParts, " / "),
			"children":   summarizeAssetGroupTree(node.Children, pathParts),
		}
		out = append(out, summary)
	}
	return out
}

func assetGroupTreeTypeLabel(nodeType int) string {
	switch nodeType {
	case 1:
		return "asset_group"
	case 2:
		return "asset"
	default:
		return "unknown"
	}
}

func firstAssetGroupSummary(groups []map[string]any) map[string]any {
	if len(groups) == 0 {
		return nil
	}
	root := groups[0]
	return map[string]any{
		"id":    root["id"],
		"name":  root["name"],
		"count": root["count"],
		"path":  root["path"],
	}
}

func assetListFilters(opts assetListOptions) map[string]any {
	filters := map[string]any{}
	if opts.id != 0 {
		filters["id"] = opts.id
	}
	if value := strings.TrimSpace(opts.name); value != "" {
		filters["name"] = value
	}
	if value := strings.TrimSpace(opts.ip); value != "" {
		filters["ip"] = value
	}
	if value := strings.TrimSpace(opts.mac); value != "" {
		filters["mac"] = value
	}
	if value := strings.TrimSpace(opts.assetType); value != "" {
		filters["asset_type"] = value
	}
	if opts.groupID != 0 {
		filters["group_id"] = opts.groupID
	}
	if value := strings.TrimSpace(opts.importance); value != "" {
		filters["importance"] = value
	}
	if value := strings.TrimSpace(opts.tagID); value != "" {
		filters["tag_id"] = value
	}
	return filters
}

func summarizeAssets(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		summary := map[string]any{}
		for _, key := range []string{
			"id",
			"name",
			"ip",
			"ip_mac",
			"importance",
			"tags",
			"contact",
			"remark",
			"asset_type",
			"location",
			"source",
			"group_info",
			"create_time",
			"update_time",
		} {
			if value, ok := item[key]; ok {
				summary[key] = value
			}
		}
		out = append(out, summary)
	}
	return out
}

func assetImportanceValue(value string) (uint, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "important", "重点", "1":
		return 1, nil
	case "normal", "普通", "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported importance %q, expected important, normal, 重点, 普通, 1, or 2", value)
	}
}

func parseUintCSV(value string) ([]uint, error) {
	parts := parseCSV(value)
	out := make([]uint, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, uint(parsed))
	}
	return out, nil
}

func parseInt64CSV(value string) ([]int64, error) {
	parts := parseCSV(value)
	out := make([]int64, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, err
		}
		if parsed <= 0 {
			return nil, fmt.Errorf("id must be greater than 0")
		}
		out = append(out, parsed)
	}
	return out, nil
}

func parseAssetCreateIPMac(value string) ([]assetCreateIPMac, error) {
	if strings.TrimSpace(value) == "" {
		return []assetCreateIPMac{}, nil
	}
	var out []assetCreateIPMac
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, fmt.Errorf("invalid ip-mac JSON: %w", err)
	}
	return out, nil
}

func parseRequiredAssetID(value string) (uint, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("missing asset id: set --id")
	}
	parsed, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid asset id: %q", value)
	}
	return uint(parsed), nil
}

func parseAssetTreeNodeType(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "group", "asset_group", "1":
		return assetTreeNodeTypeGroup, nil
	case "asset", "2":
		return assetTreeNodeTypeAsset, nil
	default:
		return 0, fmt.Errorf("expected group, asset, 1, or 2")
	}
}

func parseAssetImportResult(raw json.RawMessage) (map[string]any, error) {
	var env struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		Result struct {
			RecordCount map[string]any `json:"record_count"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("%s", env.Error.Message)
	}
	if env.Result.RecordCount == nil {
		return nil, fmt.Errorf("missing result.record_count")
	}
	return env.Result.RecordCount, nil
}

func writeAssetError(cmd *cobra.Command, task string, command string, code string, message string, retryable bool) error {
	raw, renderErr := RenderJSON(ErrorEnvelope{
		Success: false,
		Task:    task,
		Command: command,
		Error:   CLIError{Code: code, Message: message, Retryable: retryable},
	})
	if renderErr != nil {
		return renderErr
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}
