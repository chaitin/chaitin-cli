package cosmos

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

const (
	timeQueryType      = "TimeQuery"
	alarmListRPCMethod = "AlarmService.GetAlarmList"
	paramStyleArray    = "array"
)

// APIOperation 对应 JSON 中的一个 API 方法定义。
type APIOperation struct {
	Name       string              `json:"name"`
	Comment    string              `json:"comment"`
	Args       map[string]ArgField `json:"args"`
	Reply      map[string]ArgField `json:"reply"`
	ParamStyle string              `json:"paramStyle,omitempty"`
}

// ArgField 对应 JSON 中的一个参数/返回值字段。
type ArgField struct {
	Type      json.RawMessage `json:"type"`
	Optional  bool            `json:"optional"`
	TypeClass string          `json:"typeclass"`
	QueryType string          `json:"queryType,omitempty"`
	Comment   string          `json:"comment"`
}

// flagDef 描述一个需要绑定到 cobra command 上的 flag。
type flagDef struct {
	name      string // dot notation key, e.g. "filter.name"
	typeName  string // "string", "integer", "float", "bool"
	required  bool
	usage     string
	isArray   bool
	queryType string // 复杂查询类型，如 "TimeQuery", "StringQuery"
}

// parsedOp 是解析后的 API 操作，包含 service/method 信息和扁平化的 flag 列表。
type parsedOp struct {
	serviceName string // 原始 service 名，如 "AlarmService"
	methodName  string // 原始 method 名，如 "GetAlarmList"
	serviceCmd  string // kebab-case service 名，如 "alarm"
	methodCmd   string // kebab-case method 名，如 "get-alarm-list"
	comment     string
	flags       []flagDef
	paramStyle  string
	skipped     bool // 如果包含结构体数组参数则跳过
}

// parseOperations 将 JSON 解析出的 APIOperation 列表转为 parsedOp 列表。
func parseOperations(ops []APIOperation) []parsedOp {
	var result []parsedOp
	for _, op := range ops {
		p := parseOneOperation(op)
		result = append(result, p)
	}
	return result
}

func parseOneOperation(op APIOperation) parsedOp {
	parts := strings.SplitN(op.Name, ".", 2)
	if len(parts) != 2 {
		return parsedOp{skipped: true}
	}

	svcRaw := parts[0]
	methodRaw := parts[1]

	// 去掉 Service / Api 后缀
	svcShort := svcRaw
	for _, suffix := range []string{"Service", "Api"} {
		svcShort = strings.TrimSuffix(svcShort, suffix)
	}

	p := parsedOp{
		serviceName: svcRaw,
		methodName:  methodRaw,
		serviceCmd:  camelToKebab(svcShort),
		methodCmd:   camelToKebab(methodRaw),
		comment:     strings.TrimSpace(op.Comment),
		paramStyle:  op.ParamStyle,
	}

	// 解析 args 为 flags
	for key, field := range op.Args {
		if key == "-" {
			// 占位符，展开其内部字段
			expandObjectType(field.Type, "", &p)
			continue
		}
		collectFlags(key, field, &p)
	}

	return p
}

// collectFlags 将一个 ArgField 收集为 flagDef，处理嵌套和数组。
func collectFlags(prefix string, field ArgField, p *parsedOp) {
	typeName := resolveTypeName(field.Type)

	if field.TypeClass == "array" {
		if isBasicType(typeName) {
			p.flags = append(p.flags, flagDef{
				name:      prefix,
				typeName:  typeName,
				required:  !field.Optional,
				usage:     field.Comment,
				isArray:   true,
				queryType: field.QueryType,
			})
		} else {
			// 结构体数组，跳过整个 API
			p.skipped = true
		}
		return
	}

	if isBasicType(typeName) {
		p.flags = append(p.flags, flagDef{
			name:      prefix,
			typeName:  typeName,
			required:  !field.Optional,
			usage:     field.Comment,
			queryType: field.QueryType,
		})
		return
	}

	// type 是 object：展开嵌套
	expandObjectType(field.Type, prefix, p)
}

// expandObjectType 尝试将 type（JSON object）展开为嵌套 flag。
func expandObjectType(raw json.RawMessage, prefix string, p *parsedOp) {
	var obj map[string]ArgField
	if err := json.Unmarshal(raw, &obj); err != nil {
		// 无法展开（可能是自定义类型名），跳过整个 API
		p.skipped = true
		return
	}

	for key, field := range obj {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		collectFlags(fullKey, field, p)
	}
}

// resolveTypeName 从 json.RawMessage 中提取类型名。
// 如果是字符串（如 "string"）直接返回；如果是 object 返回 ""。
func resolveTypeName(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return ""
}

func isBasicType(t string) bool {
	switch t {
	case "string", "integer", "float", "bool":
		return true
	}
	return false
}

// registerOperations 将 parsedOp 列表注册为 cobra 命令树。
// 按 serviceCmd 分组，每个 service 是 parent 的子命令，method 是 service 的子命令。
func registerOperations(parent *cobra.Command, ops []parsedOp) {
	serviceMap := make(map[string]*cobra.Command)
	for _, cmd := range parent.Commands() {
		if name := commandUseName(cmd); name != "" {
			serviceMap[name] = cmd
		}
	}

	for _, op := range ops {
		if op.skipped {
			continue
		}

		svcCmd, ok := serviceMap[op.serviceCmd]
		if !ok {
			svcCmd = &cobra.Command{
				Use:   op.serviceCmd,
				Short: fmt.Sprintf("%s service", op.serviceName),
			}
			serviceMap[op.serviceCmd] = svcCmd
			parent.AddCommand(svcCmd)
		}

		methodCmd := buildMethodCommand(op)
		svcCmd.AddCommand(methodCmd)
	}
}

func commandUseName(cmd *cobra.Command) string {
	fields := strings.Fields(cmd.Use)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// buildMethodCommand 为一个 parsedOp 构建 cobra.Command，绑定 flags 和 RunE。
func buildMethodCommand(op parsedOp) *cobra.Command {
	cmd := &cobra.Command{
		Use:   op.methodCmd,
		Short: op.comment,
	}

	// flagValues 保存各 flag 的值指针，用于 RunE 中取值构造请求体。
	flagValues := make(map[string]interface{})

	for _, f := range op.flags {
		bindFlag(cmd, f, flagValues)
	}

	rpcMethod := fmt.Sprintf("%s.%s", op.serviceName, op.methodName)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		params := buildRequestBody(cmd, flagValues, op.flags)
		if err := validateRequestParams(rpcMethod, params); err != nil {
			return err
		}

		raw, _ := cmd.Flags().GetBool("raw")
		return doRequestWithParamStyle(cmd, rpcMethod, params, raw, op.paramStyle)
	}

	return cmd
}

// bindFlag 将一个 flagDef 绑定到 cobra.Command 上。
func bindFlag(cmd *cobra.Command, f flagDef, values map[string]interface{}) {
	if f.queryType != "" {
		v := cmd.Flags().String(f.name, "", f.usage)
		values[f.name] = v
		if f.required {
			_ = cmd.MarkFlagRequired(f.name)
		}
		return
	}

	switch {
	case f.isArray && f.typeName == "string":
		v := cmd.Flags().StringSlice(f.name, nil, f.usage)
		values[f.name] = v
	case f.isArray && f.typeName == "integer":
		v := cmd.Flags().IntSlice(f.name, nil, f.usage)
		values[f.name] = v
	case f.isArray && f.typeName == "bool":
		v := cmd.Flags().BoolSlice(f.name, nil, f.usage)
		values[f.name] = v
	case f.typeName == "string":
		v := cmd.Flags().String(f.name, "", f.usage)
		values[f.name] = v
	case f.typeName == "integer":
		v := cmd.Flags().Int(f.name, 0, f.usage)
		values[f.name] = v
	case f.typeName == "float":
		v := cmd.Flags().Float64(f.name, 0, f.usage)
		values[f.name] = v
	case f.typeName == "bool":
		v := cmd.Flags().Bool(f.name, false, f.usage)
		values[f.name] = v
	default:
		// 未知类型，当作 string
		v := cmd.Flags().String(f.name, "", f.usage)
		values[f.name] = v
	}

	if f.required {
		_ = cmd.MarkFlagRequired(f.name)
	}
}

// buildRequestBody 从 flagValues 构造嵌套的 JSON 请求体。
// 支持 dot notation 还原嵌套结构。
// 通过 cmd.Flags().Changed(key) 区分「未设置」与「设置为零值」，
// 避免将未设置的 bool (false) 和 int (0) 发送给 API。
// flags 携带各字段的 queryType 元数据，用于将字符串值作为 JSON 嵌入请求体。
func buildRequestBody(cmd *cobra.Command, flagValues map[string]interface{}, flags []flagDef) map[string]interface{} {
	queryTypes := make(map[string]string)
	for _, f := range flags {
		if f.queryType != "" {
			queryTypes[f.name] = f.queryType
		}
	}

	result := make(map[string]interface{})

	for key, ptr := range flagValues {
		changed := cmd.Flags().Changed(key)
		val := derefFlagValue(ptr, changed)
		if val == nil {
			continue
		}

		if queryType, ok := queryTypes[key]; ok {
			if strVal, ok := val.(string); ok && strings.TrimSpace(strVal) == "" {
				continue
			}
			setNestedValue(result, key, normalizeQueryValue(queryType, val))
			continue
		}

		setNestedValue(result, key, val)
	}

	return result
}

func normalizeQueryValue(queryType string, val interface{}) interface{} {
	strVal, ok := val.(string)
	if !ok || strings.TrimSpace(strVal) == "" {
		return val
	}

	// 标记了 queryType 的字段，其字符串值优先作为 JSON 嵌入请求体。
	// 例如 --created_at '[{"oper":"eq_in","target":"0-1781501714000"}]'.
	if embedded, ok := parseJSONValue(strVal); ok {
		return embedded
	}

	if queryType == timeQueryType {
		// Cosmos 告警列表实际使用毫秒时间戳范围，例如
		// --created_at 0-1781501714000 等价于
		// --created_at '[{"oper":"eq_in","target":"0-1781501714000"}]'.
		return []map[string]string{{
			"oper":   "eq_in",
			"target": strings.TrimSpace(strVal),
		}}
	}

	return val
}

func parseJSONValue(value string) (interface{}, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, false
	}
	if !strings.HasPrefix(value, "[") && !strings.HasPrefix(value, "{") {
		return nil, false
	}

	var embedded interface{}
	if err := json.Unmarshal([]byte(value), &embedded); err != nil {
		return nil, false
	}
	return embedded, true
}

func validateRequestParams(method string, params map[string]interface{}) error {
	if method != alarmListRPCMethod {
		return nil
	}

	if isTruthy(params["exclude_time_filter"]) || isNonZero(params["workflow_id"]) {
		return nil
	}
	if hasValue(params["created_at"]) || hasValue(params["updated_at"]) {
		return nil
	}

	return fmt.Errorf("cosmos alarm get-alarm-list requires --created_at or --updated_at; use a millisecond range, for example --created_at 0-1781501714000 or --created_at '[{\"oper\":\"eq_in\",\"target\":\"0-1781501714000\"}]'")
}

func hasValue(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(v) != ""
	case []interface{}:
		return len(v) > 0
	case []map[string]string:
		return len(v) > 0
	default:
		return true
	}
}

func isTruthy(value interface{}) bool {
	v, ok := value.(bool)
	return ok && v
}

func isNonZero(value interface{}) bool {
	switch v := value.(type) {
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(v))
		return err == nil && i != 0
	default:
		return false
	}
}

// derefFlagValue 解引用 flag 值指针，返回实际值。
// changed 表示用户是否在命令行显式设置了该 flag。
// - string：显式设置时保留空字符串，写接口用它清空字段
// - bool：仅在 changed 时返回实际值（未设置时 false 与默认值无法区分）
// - int / float64：仅在 changed 时返回实际值（未设置时 0 与默认值无法区分）
// - slice：跳过未设置的空 slice；字符串数组会过滤空项
func derefFlagValue(ptr interface{}, changed bool) interface{} {
	switch v := ptr.(type) {
	case *string:
		if v == nil {
			return nil
		}
		if *v == "" && !changed {
			return nil
		}
		return *v
	case *int:
		if v == nil {
			return nil
		}
		if !changed {
			return nil
		}
		return *v
	case *float64:
		if v == nil {
			return nil
		}
		if !changed {
			return nil
		}
		return *v
	case *bool:
		if v == nil {
			return nil
		}
		if !changed {
			return nil
		}
		return *v
	case *[]string:
		if v == nil {
			return nil
		}
		values := make([]string, 0, len(*v))
		for _, item := range *v {
			if item != "" {
				values = append(values, item)
			}
		}
		if len(values) == 0 && !changed {
			return nil
		}
		return values
	case *[]int:
		if v == nil || len(*v) == 0 {
			return nil
		}
		return *v
	case *[]bool:
		if v == nil || len(*v) == 0 {
			return nil
		}
		return *v
	}
	return nil
}

// setNestedValue 按 dot notation 设置嵌套 map 的值。
// 如 setNestedValue(m, "a.b.c", 1) → m["a"]["b"]["c"] = 1
func setNestedValue(m map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")
	current := m
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		if next, ok := current[part]; ok {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				// 冲突，覆盖
				newMap := make(map[string]interface{})
				current[part] = newMap
				current = newMap
			}
		} else {
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}
}

// camelToKebab 将 CamelCase 转换为 kebab-case。
// 例：GetAlarmList → get-alarm-list, SearchLogList → search-log-list
func camelToKebab(s string) string {
	var result []rune
	runes := []rune(s)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) {
					// aB → a-b
					result = append(result, '-')
				} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					// ABc → a-bc (处理连续大写后跟小写)
					result = append(result, '-')
				}
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}
