package cloudatlas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	api     *OpenAPI
	schemas map[string]Schema
}

func ParseSpec(data []byte) (*OpenAPI, error) {
	var api OpenAPI
	if err := yaml.Unmarshal(data, &api); err != nil {
		return nil, fmt.Errorf("parse Cloud Atlas OpenAPI spec: %w", err)
	}
	return &api, nil
}

func NewParser(api *OpenAPI) *Parser {
	applySpecOverrides(api)
	return &Parser{api: api, schemas: api.Components.Schemas}
}

// resolveSchema follows $ref chains into components.schemas. It returns the
// input schema unchanged when the reference is missing or cyclic.
func (p *Parser) resolveSchema(schema *Schema) *Schema {
	if schema == nil {
		return nil
	}
	seen := make(map[string]struct{})
	for schema.Ref != "" {
		name := refName(schema.Ref)
		if name == "" {
			return schema
		}
		if _, ok := seen[name]; ok {
			return schema
		}
		seen[name] = struct{}{}
		target, ok := p.schemas[name]
		if !ok {
			return schema
		}
		schema = &target
	}
	return schema
}

func refName(ref string) string {
	if idx := strings.LastIndex(ref, "/"); idx >= 0 {
		return ref[idx+1:]
	}
	return ref
}

func (p *Parser) GenerateCommands() ([]*cobra.Command, error) {
	parents := make(map[string]*cobra.Command)
	used := make(map[string]int)

	for _, path := range sortedPathKeys(p.api.Paths) {
		pathItem := p.api.Paths[path]
		for _, item := range operationsForPath(pathItem) {
			if item.op == nil {
				continue
			}
			if _, disabled := disabledOperations[item.method+" "+path]; disabled {
				continue
			}
			segments := p.commandPath(item.op)
			if len(segments) == 0 {
				segments = pathCommandPath(path)
			}
			leaf := actionName(item.op.Summary)
			if leaf == "" {
				leaf = strings.ToLower(item.method)
			}

			parent := getOrCreateRoot(parents, segments[0])
			current := parent
			for _, segment := range segments[1:] {
				current = getOrCreateChild(current, segment)
			}

			key := strings.Join(append(segments, leaf), " ")
			used[key]++
			use := leaf
			if used[key] > 1 {
				use = leaf + "-" + uniqueSuffix(path, item.op)
			}
			current.AddCommand(p.createOperationCommand(use, item.method, path, item.op))
		}
	}

	commands := make([]*cobra.Command, 0, len(parents))
	for _, key := range sortedCommandKeys(parents) {
		commands = append(commands, parents[key])
	}
	return commands, nil
}

type pathOperation struct {
	method string
	op     *Operation
}

func operationsForPath(pathItem PathItem) []pathOperation {
	return []pathOperation{
		{method: "GET", op: pathItem.Get},
		{method: "POST", op: pathItem.Post},
		{method: "PUT", op: pathItem.Put},
		{method: "DELETE", op: pathItem.Delete},
		{method: "PATCH", op: pathItem.Patch},
	}
}

func sortedPathKeys(paths map[string]PathItem) []string {
	keys := make([]string, 0, len(paths))
	for key := range paths {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedCommandKeys(commands map[string]*cobra.Command) []string {
	keys := make([]string, 0, len(commands))
	for key := range commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// tagCommandPaths maps the OpenAPI tag of each operation group to the command
// path it should be exposed under.
var tagCommandPaths = map[string][]string{
	"种子-企业主体":        {"seed", "enterprise"},
	"种子-域名 WHOIS":    {"seed", "domain-whois"},
	"种子-证书信息":        {"seed", "certificate-info"},
	"种子-网站图标":        {"seed", "favicon"},
	"种子-邮箱域名":        {"seed", "email-domain"},
	"种子-关键词":         {"seed", "keyword"},
	"种子-网站标题":        {"seed", "title"},
	"种子-网页哈希":        {"seed", "page-hash"},
	"资产-主域名":         {"asset", "root-domain"},
	"资产-子域名":         {"asset", "subdomain"},
	"资产-解析记录":        {"asset", "dns-record"},
	"资产-IP 地址":       {"asset", "ip"},
	"资产-证书":          {"asset", "certificate"},
	"资产-解析记录(V2)":    {"asset", "dns-record-v2"},
	"资产-IP 地址(V2)":   {"asset", "ip-v2"},
	"暴露面-端口服务":       {"exposure", "port-service"},
	"暴露面-网站实体":       {"exposure", "website"},
	"暴露面-网站路径":       {"exposure", "website-path"},
	"暴露面-网站指纹":       {"exposure", "website-fingerprint"},
	"暴露面-爬虫数据":       {"exposure", "crawler"},
	"暴露面-开放端口":       {"exposure", "open-port"},
	"风险-漏洞列表":        {"risk", "vulnerability"},
	"风险-高危端口":        {"risk", "high-risk-port"},
	"风险-高危应用":        {"risk", "high-risk-app"},
	"风险-管理后台":        {"risk", "admin-panel"},
	"数字情报-开发社区泄露":    {"intelligence", "code-leak"},
	"数字情报-网盘泄露":      {"intelligence", "cloud-drive-leak"},
	"数字情报-文库泄露":      {"intelligence", "doc-leak"},
	"数字情报-暗网情报":      {"intelligence", "dark-web"},
	"数字情报-企业邮箱":      {"intelligence", "enterprise-email"},
	"数字情报-失窃数据":      {"intelligence", "stolen-data"},
	"数字情报-移动应用":      {"intelligence", "mobile-app"},
	"数字情报-新媒体":       {"intelligence", "new-media"},
	"数字情报-监控规则":      {"intelligence", "monitor-rule"},
	"数字情报-排除策略":      {"intelligence", "exclude-policy"},
	"数字情报-监控任务":      {"intelligence", "monitor-task"},
	"任务管理-任务队列":      {"task", "queue"},
	"任务管理-资产发现任务实例":  {"task", "asset", "instance"},
	"任务管理-暴露面识别任务实例": {"task", "exposure", "instance"},
	"任务管理-风险评估任务实例":  {"task", "risk", "instance"},
	"任务管理-资产发现周期任务":  {"task", "asset", "schedule"},
	"任务管理-暴露面识别周期任务": {"task", "exposure", "schedule"},
	"任务管理-风险评估周期任务":  {"task", "risk", "schedule"},
	"任务管理-资产发现任务剧本":  {"task", "asset", "flow"},
	"任务管理-暴露面识别任务剧本": {"task", "exposure", "flow"},
	"任务管理-风险评估任务剧本":  {"task", "risk", "flow"},
	"空间管理-空间":        {"space"},
	"空间管理-资产标签":      {"space", "tag"},
	"空间管理-资产分组":      {"space", "group"},
	"空间管理-扫描节点":      {"space", "scanner-node"},
	"系统管理-扫描集群":      {"system", "scanner-cluster"},
	"用户管理-空间成员":      {"user", "member"},
	"策略-厂商":          {"strategy", "vendor"},
	"策略-产品指纹":        {"strategy", "product-fingerprint"},
	"策略-漏洞分组":        {"strategy", "vulnerability-group"},
	"策略-漏洞列表":        {"strategy", "vulnerability"},
}

// specOverrides records corrections for gaps in the upstream Cloud Atlas
// OpenAPI document. The document is a vendor snapshot, so re-exporting it would
// silently drop any edit made directly to the file; corrections the server
// actually enforces therefore live here instead. Keys are "<METHOD> <path>"
// exactly as they appear in the document.
type specOverride struct {
	// requiredQueryParameters lists query parameters the server rejects when
	// missing but the document does not mark as required.
	requiredQueryParameters []string
	// bodyProperties patches request body fields, keyed by property name.
	bodyProperties map[string]propertyOverride
}

type propertyOverride struct {
	description string
	enum        []any
}

var specOverrides = map[string]specOverride{
	"GET /v1/task/asset/flow":  {requiredQueryParameters: []string{"task_id"}},
	"GET /v1/task/attack/flow": {requiredQueryParameters: []string{"task_id"}},
	"GET /v1/task/risk/flow":   {requiredQueryParameters: []string{"task_id"}},
	"POST /v1/space/space/{id}/truncate": {
		bodyProperties: map[string]propertyOverride{
			"resources": {
				description: "要清空的资源类型；服务端可选项: attack, seed, risk, drps, asset",
				enum:        []any{"attack", "seed", "risk", "drps", "asset"},
			},
		},
	},
}

func applySpecOverrides(api *OpenAPI) {
	if api == nil {
		return
	}
	for key, override := range specOverrides {
		method, path, ok := strings.Cut(key, " ")
		if !ok {
			continue
		}
		op := operationFor(api.Paths[path], method)
		if op == nil {
			continue
		}
		for _, name := range override.requiredQueryParameters {
			for i := range op.Parameters {
				if op.Parameters[i].In == "query" && op.Parameters[i].Name == name {
					op.Parameters[i].Required = true
				}
			}
		}
		if len(override.bodyProperties) > 0 {
			applyBodyOverrides(api, op, override.bodyProperties)
		}
	}
}

func operationFor(pathItem PathItem, method string) *Operation {
	for _, item := range operationsForPath(pathItem) {
		if item.method == method {
			return item.op
		}
	}
	return nil
}

// disabledOperations temporarily hides operations whose upstream API does not
// work, so the CLI does not advertise commands that can never succeed. Keys use
// the same "<METHOD> <path>" form as specOverrides and the value records why the
// operation is hidden. Delete an entry once the endpoint is fixed server-side;
// no other change is needed to bring the command back.
//
// Verified 2026-09-21 against the pre-test environment: every operation below
// answers HTTP 200 with {"code":500,"message":"服务异常，请联系管理员!"} for any
// parameters. The sibling .../async/download endpoint of the same resource works,
// which points at these two handlers rather than at the request the CLI sends.
var disabledOperations = map[string]string{
	"GET /v1/drps/rule/download":                "异步下载监控规则：服务端固定 500",
	"GET /v1/risk/admin/download":               "异步下载管理后台：服务端固定 500",
	"GET /v1/risk/product/download":             "异步下载高危应用：服务端固定 500",
	"GET /v1/risk/service/download":             "异步下载高危端口：服务端固定 500",
	"GET /v1/drps/rule/{id}/download/status":    "监控规则下载状态：服务端固定 500",
	"GET /v1/risk/admin/{id}/download/status":   "管理后台下载状态：服务端固定 500",
	"GET /v1/risk/product/{id}/download/status": "高危应用下载状态：服务端固定 500",
	"GET /v1/risk/service/{id}/download/status": "高危端口下载状态：服务端固定 500",
}

// applyBodyOverrides patches the JSON request body schema of op. The schema is
// usually a $ref into components.schemas, so the patched copy is written back
// to the component map.
func applyBodyOverrides(api *OpenAPI, op *Operation, patches map[string]propertyOverride) {
	schema, ref := requestBodySchema(api, op)
	if schema == nil {
		return
	}
	for name, patch := range patches {
		prop, ok := schema.Properties[name]
		if !ok {
			continue
		}
		if patch.description != "" {
			prop.Description = patch.description
		}
		if len(patch.enum) > 0 {
			prop.Enum = patch.enum
		}
		schema.Properties[name] = prop
	}
	if ref != "" {
		api.Components.Schemas[ref] = *schema
	}
}

func requestBodySchema(api *OpenAPI, op *Operation) (*Schema, string) {
	if op.RequestBody == nil {
		return nil, ""
	}
	media, ok := op.RequestBody.Content["application/json"]
	if !ok || media.Schema == nil {
		return nil, ""
	}
	if media.Schema.Ref == "" {
		return media.Schema, ""
	}
	ref := refName(media.Schema.Ref)
	target, ok := api.Components.Schemas[ref]
	if !ok {
		return nil, ""
	}
	return &target, ref
}

func (p *Parser) commandPath(op *Operation) []string {
	if len(op.Tags) == 0 || strings.TrimSpace(op.Tags[0]) == "" {
		return nil
	}
	tag := strings.TrimSpace(op.Tags[0])
	if segments, ok := tagCommandPaths[tag]; ok {
		return segments
	}
	segments := make([]string, 0, 3)
	for _, part := range strings.Split(tag, "/") {
		for _, piece := range strings.Split(part, "-") {
			if segment := normalizeSegment(piece); segment != "" {
				segments = append(segments, segment)
			}
		}
	}
	return segments
}

func pathCommandPath(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "{") || part == "v1" || part == "v2" || part == "api" {
			continue
		}
		segments = append(segments, normalizeSegment(part))
	}
	return segments
}

type actionRule struct {
	contains []string
	name     string
}

// actionRules translate operation summaries into stable subcommand names. The
// list is ordered from most to least specific; the first rule whose keywords
// are all present in the summary wins.
var actionRules = []actionRule{
	{[]string{"快速检测任务", "删除"}, "delete-task"},
	{[]string{"快速检测任务", "重跑"}, "rerun-task"},
	{[]string{"快速检测任务", "下发"}, "create-task"},
	{[]string{"快速检测任务", "列表"}, "task-list"},
	{[]string{"检测任务", "列表"}, "task-list"},
	{[]string{"同步下载"}, "sync-download"},
	{[]string{"异步下载"}, "async-download"},
	{[]string{"下载状态"}, "download-status"},
	{[]string{"应用指纹任务状态"}, "task-status"},
	{[]string{"应用指纹"}, "finger"},
	{[]string{"定时任务配置", "更新"}, "update-schedule"},
	{[]string{"定时任务配置", "查询"}, "get-schedule"},
	{[]string{"来源选项"}, "source-options"},
	{[]string{"状态选项"}, "status-options"},
	{[]string{"标签选项"}, "tag-options"},
	{[]string{"分组选项"}, "group-options"},
	{[]string{"主机名选项"}, "hostname-options"},
	{[]string{"关联产品"}, "related-products"},
	{[]string{"关联端口"}, "related-ports"},
	{[]string{"状态数量"}, "status-count"},
	{[]string{"状态统计"}, "status-count"},
	{[]string{"统计"}, "statistic"},
	{[]string{"选项"}, "options"},
	{[]string{"概览"}, "overview"},
	{[]string{"聚合数据"}, "aggregate"},
	{[]string{"批量绑定"}, "bind-tags"},
	{[]string{"批量解绑"}, "unbind-tags"},
	{[]string{"全量更新", "标签"}, "update-tags"},
	{[]string{"批量修改"}, "bind-group"},
	{[]string{"批量更新"}, "set-status"},
	{[]string{"批量切换"}, "set-status"},
	{[]string{"更新置信度"}, "set-confidence"},
	{[]string{"批量启停"}, "set-enabled"},
	{[]string{"按状态", "删除"}, "delete-by-status"},
	{[]string{"批量标记删除"}, "mark-deleted"},
	{[]string{"批量删除"}, "batch-delete"},
	{[]string{"删除"}, "delete"},
	{[]string{"批量重启"}, "reboot"},
	{[]string{"漏洞复测"}, "recheck"},
	{[]string{"标记子域名为人工录入"}, "set-manual"},
	{[]string{"手动触发"}, "trigger"},
	{[]string{"清空"}, "truncate"},
	{[]string{"密码"}, "update-password"},
	{[]string{"部分更新"}, "patch"},
	{[]string{"创建"}, "create"},
	{[]string{"更新"}, "update"},
	{[]string{"查询", "列表"}, "list"},
	{[]string{"查询", "详情"}, "get"},
	{[]string{"查询"}, "get"},
}

func actionName(summary string) string {
	summary = strings.TrimSpace(summary)
	for _, rule := range actionRules {
		matched := true
		for _, keyword := range rule.contains {
			if !strings.Contains(summary, keyword) {
				matched = false
				break
			}
		}
		if matched {
			return rule.name
		}
	}
	return normalizeSegment(summary)
}

func getOrCreateRoot(parents map[string]*cobra.Command, name string) *cobra.Command {
	if parents[name] != nil {
		return parents[name]
	}
	cmd := &cobra.Command{Use: name, Short: fmt.Sprintf("%s commands", name)}
	parents[name] = cmd
	return cmd
}

func getOrCreateChild(parent *cobra.Command, name string) *cobra.Command {
	for _, child := range parent.Commands() {
		if child.Name() == name {
			return child
		}
	}
	cmd := &cobra.Command{Use: name, Short: fmt.Sprintf("%s commands", name)}
	parent.AddCommand(cmd)
	return cmd
}

func (p *Parser) createOperationCommand(use, method, path string, op *Operation) *cobra.Command {
	short := strings.TrimSpace(op.Summary)
	if short == "" {
		short = fmt.Sprintf("%s %s", method, path)
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  p.buildOperationHelp(method, path, op),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.executeCommand(cmd, method, path, op)
		},
	}

	for _, param := range op.Parameters {
		if param.In != "query" && param.In != "path" {
			continue
		}
		flagName := parameterFlagName(param)
		required := param.Required || flagName == "space"
		cmd.Flags().String(flagName, "", p.formatParameterUsage(param, required))
	}
	if !hasQueryParameter(op.Parameters, "space") {
		cmd.Flags().String("space", "", "空间ID；必填；可由 --space-id 或 cloudAtlas.space_id 提供默认值")
	}
	cmd.Flags().StringArray("extra-query", nil, "附加 query 参数，格式 key=value，可重复；用于传入 OpenAPI 未建模的查询参数")
	if op.RequestBody != nil {
		cmd.Flags().String("body", "", "JSON request body；字段说明见本命令 help 的 Body 区域")
		cmd.Flags().String("body-file", "", "读取 JSON request body 的文件路径；字段说明见本命令 help 的 Body 区域")
	}
	if requiresConfirmation(method, op.Summary) {
		cmd.Flags().Bool("yes", false, "Confirm this modifying operation")
	}
	return cmd
}

func (p *Parser) executeCommand(cmd *cobra.Command, method, path string, op *Operation) error {
	if requiresConfirmation(method, op.Summary) {
		yes, err := cmd.Flags().GetBool("yes")
		if err != nil {
			return err
		}
		if !yes {
			return fmt.Errorf("operation %q requires --yes", op.Summary)
		}
	}

	requestPath, err := buildPath(cmd, path, op.Parameters)
	if err != nil {
		return err
	}
	query, err := buildQuery(cmd, op.Parameters)
	if err != nil {
		return err
	}
	body, err := buildBody(cmd, op)
	if err != nil {
		return err
	}

	cfg := getConfigFromCommand(cmd)
	client := NewClient(cfg, runtimeInsecure, verbose)
	data, err := client.Do(cmd.Context(), method, requestPath, query, body)
	if err != nil {
		if _, ok := err.(dryRunResult); !ok {
			return err
		}
	}
	return NewRenderer(outputFormat, cmd.OutOrStdout()).Render(data)
}

func buildPath(cmd *cobra.Command, path string, params []Parameter) (string, error) {
	result := path
	for _, param := range params {
		if param.In != "path" {
			continue
		}
		flagName := parameterFlagName(param)
		value, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return "", err
		}
		if value == "" && param.Required {
			return "", fmt.Errorf("--%s is required", flagName)
		}
		result = strings.ReplaceAll(result, "{"+param.Name+"}", url.PathEscape(value))
	}
	return result, nil
}

func buildQuery(cmd *cobra.Command, params []Parameter) (url.Values, error) {
	query := url.Values{}
	spaceHandled := false
	for _, param := range params {
		if param.In != "query" {
			continue
		}
		flagName := parameterFlagName(param)
		if flagName == "space" {
			spaceHandled = true
		}
		value, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return nil, err
		}
		if flagName == "space" && value == "" {
			value = spaceIDFromCommand(cmd)
		}
		required := param.Required || flagName == "space"
		if value == "" {
			if required {
				return nil, fmt.Errorf("--%s is required", flagName)
			}
			continue
		}
		query.Set(param.Name, value)
	}
	if !spaceHandled {
		space, err := cmd.Flags().GetString("space")
		if err != nil {
			return nil, err
		}
		if space == "" {
			space = spaceIDFromCommand(cmd)
		}
		if space == "" {
			return nil, fmt.Errorf("--space or --space-id is required")
		}
		query.Set("space", space)
	}
	pairs, err := cmd.Flags().GetStringArray("extra-query")
	if err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid --extra-query %q, expected key=value", pair)
		}
		query.Add(key, value)
	}
	return query, nil
}

func hasQueryParameter(params []Parameter, name string) bool {
	for _, param := range params {
		if param.In == "query" && parameterFlagName(param) == name {
			return true
		}
	}
	return false
}

func buildBody(cmd *cobra.Command, op *Operation) (any, error) {
	if op.RequestBody == nil {
		return nil, nil
	}
	bodyText, err := cmd.Flags().GetString("body")
	if err != nil {
		return nil, err
	}
	bodyFile, err := cmd.Flags().GetString("body-file")
	if err != nil {
		return nil, err
	}
	if bodyText != "" && bodyFile != "" {
		return nil, fmt.Errorf("--body and --body-file cannot be used together")
	}
	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("read body file: %w", err)
		}
		bodyText = string(data)
	}
	if bodyText == "" {
		return nil, nil
	}
	var body any
	if err := json.Unmarshal([]byte(bodyText), &body); err != nil {
		return nil, fmt.Errorf("parse body JSON: %w", err)
	}
	return body, nil
}

func getConfigFromCommand(cmd *cobra.Command) Config {
	cfg := runtimeCfg
	if value, err := cmd.Flags().GetString("url"); err == nil && value != "" {
		cfg.URL = value
	}
	if value, err := cmd.Flags().GetString("token"); err == nil && value != "" {
		cfg.Token = value
	}
	if value := spaceIDFromCommand(cmd); value != "" {
		cfg.SpaceID = value
	}
	return cfg
}

func spaceIDFromCommand(cmd *cobra.Command) string {
	if value, err := cmd.Flags().GetString("space-id"); err == nil && value != "" {
		return value
	}
	if value, err := cmd.InheritedFlags().GetString("space-id"); err == nil && value != "" {
		return value
	}
	return runtimeCfg.SpaceID
}

func requiresConfirmation(method, summary string) bool {
	if method == "DELETE" {
		return true
	}
	for _, keyword := range []string{
		"批量删除", "批量标记删除", "批量启停", "批量更新", "批量切换", "批量绑定", "批量解绑",
		"批量修改", "批量重启", "全量更新", "更新置信度", "重跑", "漏洞复测", "手动触发", "下发",
		"清空", "密码", "终止", "解绑",
	} {
		if strings.Contains(summary, keyword) {
			return true
		}
	}
	return false
}

func (p *Parser) buildOperationHelp(method, path string, op *Operation) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\nEndpoint: %s %s\n", op.Summary, method, path)
	if op.Description != "" {
		fmt.Fprintf(&b, "\n%s\n", op.Description)
	}
	if len(op.Parameters) > 0 {
		fmt.Fprintln(&b, "\nParameters:")
		for _, param := range op.Parameters {
			required := param.Required || parameterFlagName(param) == "space"
			fmt.Fprintf(&b, "  --%s (%s): %s\n", parameterFlagName(param), param.In, p.formatParameterUsage(param, required))
		}
	}
	if op.RequestBody != nil {
		fmt.Fprintln(&b, "\nBody: use --body or --body-file with JSON input.")
		bodyHelp := p.formatRequestBodyHelp(op.RequestBody)
		if bodyHelp != "" {
			fmt.Fprintf(&b, "%s\n", bodyHelp)
		}
	}
	return strings.TrimSpace(b.String())
}

func (p *Parser) formatParameterUsage(param Parameter, required bool) string {
	resolved := param
	resolved.Schema = p.resolveSchema(param.Schema)
	return formatParameterUsage(resolved, required)
}

func formatParameterUsage(param Parameter, required bool) string {
	parts := make([]string, 0, 7)
	description := strings.TrimSpace(param.Description)
	if description == "" {
		description = "OpenAPI 未提供说明"
	}
	parts = append(parts, description)
	if required {
		parts = append(parts, "必填")
	}
	parts = append(parts, formatSchemaDetails(param.Schema, param.Example)...)
	if normalizeFlagName(param.Name) == "space" {
		parts = append(parts, "可由 --space-id 或 cloudAtlas.space_id 提供默认值")
	} else if parameterFlagName(param) != normalizeFlagName(param.Name) {
		parts = append(parts, "对应 OpenAPI 参数名: "+param.Name)
	}
	return strings.Join(parts, "；")
}

func formatSchemaDetails(schema *Schema, example any) []string {
	if schema == nil {
		return nil
	}
	parts := make([]string, 0, 4)
	if typ := schemaTypeName(schema); typ != "" {
		parts = append(parts, "类型: "+typ)
	}
	if defaultText := valueHelp(schema.Default); defaultText != "" {
		parts = append(parts, "默认值: "+defaultText)
	}
	if exampleText := valueHelp(firstNonNil(example, schema.Example)); exampleText != "" {
		parts = append(parts, "示例: "+exampleText)
	}
	if enumText := formatEnumValues(schema); enumText != "" {
		parts = append(parts, "可选值: "+enumText)
	}
	return parts
}

func schemaTypeName(schema *Schema) string {
	if schema == nil {
		return ""
	}
	if schema.Type != "" {
		if schema.Type == "array" && schema.Items != nil {
			if inner := schemaTypeName(schema.Items); inner != "" {
				return "array<" + inner + ">"
			}
		}
		if schema.Nullable {
			return schema.Type + "|null"
		}
		return schema.Type
	}
	if schema.Ref != "" {
		return refName(schema.Ref)
	}
	if len(schema.Properties) > 0 {
		return "object"
	}
	if schema.Items != nil {
		return "array"
	}
	return ""
}

func formatEnumValues(schema *Schema) string {
	if schema == nil || len(schema.Enum) == 0 {
		return ""
	}
	descriptions := enumDescriptions(schema)
	items := make([]string, 0, len(schema.Enum))
	for _, value := range schema.Enum {
		valueText := valueHelp(value)
		if valueText == "" {
			continue
		}
		if description := strings.TrimSpace(descriptions[valueText]); description != "" {
			items = append(items, valueText+"="+description)
			continue
		}
		items = append(items, valueText)
	}
	return strings.Join(items, ", ")
}

func enumDescriptions(schema *Schema) map[string]string {
	descriptions := make(map[string]string)
	if schema == nil {
		return descriptions
	}
	for value, description := range schema.XApifox.EnumDescriptions {
		if strings.TrimSpace(description) != "" {
			descriptions[value] = strings.TrimSpace(description)
		}
	}
	for _, item := range schema.XApifoxEnum {
		value := valueHelp(item.Value)
		if value == "" {
			continue
		}
		description := strings.TrimSpace(item.Description)
		if description == "" {
			description = strings.TrimSpace(item.Name)
		}
		if description != "" {
			descriptions[value] = description
		}
	}
	return descriptions
}

func (p *Parser) formatRequestBodyHelp(body *RequestBody) string {
	if body == nil {
		return ""
	}
	media := body.Content["application/json"]
	if media.Schema == nil {
		for _, candidate := range body.Content {
			media = candidate
			break
		}
	}
	lines := p.formatSchemaFields(media.Schema, "  ")
	if len(lines) == 0 {
		return "  OpenAPI 未提供可展开的 body 字段说明。"
	}
	return strings.Join(lines, "\n")
}

func (p *Parser) formatSchemaFields(schema *Schema, indent string) []string {
	resolved := p.resolveSchema(schema)
	if resolved == nil || len(resolved.Properties) == 0 {
		return nil
	}
	required := make(map[string]struct{}, len(resolved.Required))
	for _, name := range resolved.Required {
		required[name] = struct{}{}
	}
	names := make([]string, 0, len(resolved.Properties))
	for name := range resolved.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	lines := make([]string, 0, len(names))
	for _, name := range names {
		prop := resolved.Properties[name]
		resolvedProp := p.resolveSchema(&prop)
		parts := make([]string, 0, 6)
		if description := strings.TrimSpace(resolvedProp.Description); description != "" {
			parts = append(parts, description)
		} else if description := strings.TrimSpace(prop.Description); description != "" {
			parts = append(parts, description)
		} else {
			parts = append(parts, "OpenAPI 未提供说明")
		}
		if _, ok := required[name]; ok {
			parts = append(parts, "必填")
		}
		parts = append(parts, formatSchemaDetails(resolvedProp, resolvedProp.Example)...)
		lines = append(lines, fmt.Sprintf("%s%s: %s", indent, name, strings.Join(parts, "；")))
	}
	return lines
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func valueHelp(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return strings.TrimSpace(fmt.Sprint(v))
		}
		return strings.TrimSpace(string(data))
	}
}

// reservedFlagNames are persistent or generated flags that an operation
// parameter must not shadow.
var reservedFlagNames = map[string]struct{}{
	"url":               {},
	"token":             {},
	"space-id":          {},
	"output":            {},
	"insecure":          {},
	"verbose":           {},
	"verbose-sensitive": {},
	"yes":               {},
	"body":              {},
	"body-file":         {},
	"extra-query":       {},
	"help":              {},
}

func parameterFlagName(param Parameter) string {
	name := normalizeFlagName(param.Name)
	if name == "" || name == "space" {
		return name
	}
	if _, reserved := reservedFlagNames[name]; reserved {
		return name + "-filter"
	}
	return name
}

func normalizeFlagName(name string) string {
	return strings.ReplaceAll(normalizeSegment(name), "_", "-")
}

func normalizeSegment(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ToLower(value)

	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if r == '-' && !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

var pathParamPattern = regexp.MustCompile(`\{([^}]+)\}`)

func uniqueSuffix(path string, op *Operation) string {
	cleaned := pathParamPattern.ReplaceAllString(strings.Trim(path, "/"), "$1")
	parts := strings.Split(cleaned, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		segment := normalizeSegment(parts[i])
		if segment == "" || segment == "v1" || segment == "v2" || segment == "api" {
			continue
		}
		return segment
	}
	if op.OperationID != "" {
		return normalizeSegment(op.OperationID)
	}
	return "operation"
}
