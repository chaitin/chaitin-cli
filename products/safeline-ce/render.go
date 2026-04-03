package safelinece

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"
)

// OutputFormat 输出格式
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
)

// Renderer 渲染接口
type Renderer interface {
	Render(data interface{}) error
}

// TableRenderer 表格渲染器
type TableRenderer struct {
	out io.Writer
}

// JSONRenderer JSON 渲染器
type JSONRenderer struct {
	out io.Writer
}

// NewRenderer 创建渲染器
func NewRenderer(format OutputFormat, out io.Writer) Renderer {
	switch format {
	case FormatJSON:
		return &JSONRenderer{out: out}
	default:
		return &TableRenderer{out: out}
	}
}

// Render 实现 JSON 渲染
func (r *JSONRenderer) Render(data interface{}) error {
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// Render 实现表格渲染
func (r *TableRenderer) Render(data interface{}) error {
	// 从 APIResponse 提取 data 字段
	extracted := extractData(data)

	if extracted == nil {
		fmt.Fprintln(r.out, "No data found")
		return nil
	}

	val := reflect.ValueOf(extracted)
	if val.Kind() == reflect.Slice {
		return r.renderSlice(val)
	}
	return r.renderSingle(val)
}

// 重要字段优先级（按顺序显示）
var importantFields = map[string][]string{
	"default": {"id", "name", "status", "state", "enabled", "created_at", "updated_at"},
	"site":    {"id", "server_names", "ports", "upstreams", "state", "comment"},
	"rule":    {"id", "name", "is_enabled", "action", "pattern", "comment"},
	"ipgroup": {"id", "comment", "ips"},
	"attack":  {"id", "host", "url", "ip", "attack_type", "action", "created_at"},
	"audit":   {"id", "username", "content", "ip", "created_at"},
	"stat":    {"time", "value"},
}

func (r *TableRenderer) renderSlice(val reflect.Value) error {
	if val.Len() == 0 {
		fmt.Fprintln(r.out, "No data found")
		return nil
	}

	// 获取第一行数据
	first := val.Index(0).Interface()
	firstMap, ok := first.(map[string]interface{})
	if !ok {
		// 非 map 类型，使用默认渲染
		return r.renderSliceDefault(val)
	}

	// 选择要显示的列
	columns := selectColumns(firstMap)

	// 打印表头
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = formatColumnName(col)
	}
	r.printRow(headers)

	// 打印分隔线
	separators := make([]string, len(headers))
	for i, h := range headers {
		separators[i] = strings.Repeat("-", max(len(h), 4))
	}
	r.printRow(separators)

	// 打印数据行
	for i := 0; i < val.Len(); i++ {
		row := extractRowFromMap(val.Index(i).Interface(), columns)
		r.printRow(row)
	}

	return nil
}

func (r *TableRenderer) renderSliceDefault(val reflect.Value) error {
	columns := inferColumns(val.Index(0).Interface())
	if len(columns) == 0 {
		fmt.Fprintln(r.out, "No data found")
		return nil
	}

	r.printRow(columns)
	separators := make([]string, len(columns))
	for i, col := range columns {
		separators[i] = strings.Repeat("-", len(col))
	}
	r.printRow(separators)

	for i := 0; i < val.Len(); i++ {
		row := extractRow(val.Index(i).Interface(), columns)
		r.printRow(row)
	}

	return nil
}

func (r *TableRenderer) renderSingle(val reflect.Value) error {
	m, ok := val.Interface().(map[string]interface{})
	if ok {
		return r.renderMapAsTable(m)
	}

	columns := inferColumns(val.Interface())
	if len(columns) == 0 {
		// 单个值直接打印
		fmt.Fprintln(r.out, formatValue(val))
		return nil
	}

	r.printRow(columns)
	separators := make([]string, len(columns))
	for i, col := range columns {
		separators[i] = strings.Repeat("-", len(col))
	}
	r.printRow(separators)
	r.printRow(extractRow(val.Interface(), columns))

	return nil
}

func (r *TableRenderer) renderMapAsTable(m map[string]interface{}) error {
	// 按 key 排序
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// 打印 key-value 表格
	maxKeyLen := 0
	for _, k := range keys {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	for _, k := range keys {
		fmt.Fprintf(r.out, "%-*s  %s\n", maxKeyLen, strings.ToUpper(k), formatValueSimple(m[k]))
	}

	return nil
}

func (r *TableRenderer) printRow(row []string) {
	for i, col := range row {
		if i > 0 {
			fmt.Fprint(r.out, "\t")
		}
		fmt.Fprint(r.out, col)
	}
	fmt.Fprintln(r.out)
}

// extractData 从嵌套结构中提取数据
func extractData(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	// 尝试从 map 中提取 data 字段（可能多层嵌套）
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	// 检查是否有 data 字段
	if d, exists := m["data"]; exists {
		// 递归提取（处理嵌套的 data.data 结构）
		return extractData(d)
	}

	return m
}

// selectColumns 选择要显示的列
func selectColumns(m map[string]interface{}) []string {
	// 根据字段推断数据类型
	dataType := inferDataType(m)

	// 获取该类型的重要字段
	priorityFields := importantFields[dataType]
	if priorityFields == nil {
		priorityFields = importantFields["default"]
	}

	// 优先显示重要字段
	var columns []string
	added := make(map[string]bool)

	// 先添加重要字段
	for _, field := range priorityFields {
		if _, exists := m[field]; exists {
			columns = append(columns, field)
			added[field] = true
		}
	}

	// 再添加其他字段（跳过一些不重要的字段）
	skipFields := map[string]bool{
		"pattern":            true,
		"auth_source_ids":    true,
		"cloud_id":           true,
		"cloud_total":        true,
		"compatible":         true,
		"builtin":            true,
		"negate":             true,
		"replay":             true,
		"review":             true,
		"tfa_enabled":        true,
		"auth_callback":      true,
		"auth_rule":          true,
		"black_rule":         true,
		"white_rule":         true,
		"captcha_rule":       true,
		"pass_count":         true,
		"req_count":          true,
		"expire":             true,
		"level":              true,
		"log":                true,
		"init":               true,
		"cert_id":            true,
		"health_check":       true,
		"exclude_paths":      true,
		"forbidden_status_code": true,
		"not_found_status_code": true,
		"acl_response_html_path": true,
		"index":              true,
		"load_balance":       true,
		"redirect_status_code": true,
		"gateway_timeout_html_path": true,
		"gateway_timeout_status_code": true,
		"bad_gateway_html_path": true,
		"bad_gateway_status_code": true,
		"chaos_id":            true,
		"chaos_is_enabled":    true,
		"custom_location":     true,
		"access_log_limit":    true,
		"error_log_limit":     true,
		"group_id":            true,
		"email":              true,
		"icon":               true,
		"ssl":                true,
		"stat_enabled":       true,
		"sp_enabled":         true,
		"static_default":     true,
		"user_agent":         true,
		"client_max_body_size": true,
		"cache":              true,
		"cache_ttl":          true,
		"retry":              true,
		"owner":              true,
		"state":              true,
		"healthy":            true,
		"health_state":       true,
		"start_time":         true,
		"end_time":           true,
	}

	// 获取所有字段并排序
	allFields := make([]string, 0, len(m))
	for k := range m {
		allFields = append(allFields, k)
	}
	sort.Strings(allFields)

	for _, field := range allFields {
		if added[field] {
			continue
		}
		if skipFields[field] {
			continue
		}
		columns = append(columns, field)
	}

	// 限制列数
	if len(columns) > 10 {
		columns = columns[:10]
	}

	return columns
}

// inferDataType 根据字段推断数据类型
func inferDataType(m map[string]interface{}) string {
	// 检查特征字段
	if _, ok := m["server_names"]; ok {
		return "site"
	}
	if _, ok := m["pattern"]; ok {
		return "rule"
	}
	if _, ok := m["ips"]; ok {
		return "ipgroup"
	}
	return "default"
}

// formatColumnName 格式化列名
func formatColumnName(name string) string {
	return strings.ToUpper(name)
}

// extractRowFromMap 从 map 提取行数据
func extractRowFromMap(v interface{}, columns []string) []string {
	row := make([]string, len(columns))

	m, ok := v.(map[string]interface{})
	if !ok {
		return row
	}

	for i, col := range columns {
		if val, exists := m[col]; exists {
			row[i] = formatValueSimple(val)
		}
	}

	return row
}

func inferColumns(v interface{}) []string {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct && val.Kind() != reflect.Map {
		return nil
	}

	var columns []string

	if val.Kind() == reflect.Map {
		for _, key := range val.MapKeys() {
			columns = append(columns, strings.ToUpper(key.String()))
		}
	} else {
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				name := strings.Split(jsonTag, ",")[0]
				if name != "" {
					columns = append(columns, strings.ToUpper(name))
				}
			} else {
				columns = append(columns, strings.ToUpper(field.Name))
			}
		}
	}

	return columns
}

func extractRow(v interface{}, columns []string) []string {
	row := make([]string, len(columns))
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i, col := range columns {
		fieldName := strings.ToLower(col)
		row[i] = getFieldValue(val, fieldName)
	}

	return row
}

func getFieldValue(val reflect.Value, fieldName string) string {
	if val.Kind() == reflect.Map {
		for _, key := range val.MapKeys() {
			if strings.ToLower(key.String()) == fieldName {
				v := val.MapIndex(key)
				return formatValueSimple(v.Interface())
			}
		}
		return ""
	}

	if val.Kind() == reflect.Struct {
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			name := field.Name
			if jsonTag != "" && jsonTag != "-" {
				name = strings.Split(jsonTag, ",")[0]
			}
			if strings.ToLower(name) == fieldName {
				return formatValueSimple(val.Field(i).Interface())
			}
		}
	}

	return ""
}

// formatValueSimple 简化的值格式化
func formatValueSimple(v interface{}) string {
	if v == nil {
		return ""
	}

	val := reflect.ValueOf(v)
	return formatValue(val)
}

func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	// 处理接口类型
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		// 截断长字符串
		if len(s) > 50 {
			return s[:47] + "..."
		}
		return s
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 检查是否是时间戳
		if v.Type().Name() == "" || v.Type().Name() == "int64" {
			// 尝试作为时间戳解析（毫秒）
			ts := v.Int()
			if ts > 1e12 && ts < 2e12 {
				// 可能是毫秒时间戳
				t := time.Unix(ts/1000, 0)
				return t.Format("2006-01-02 15:04:05")
			} else if ts > 1e9 && ts < 2e9 {
				// 可能是秒时间戳
				t := time.Unix(ts, 0)
				return t.Format("2006-01-02 15:04:05")
			}
		}
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		// 检查是否是时间戳（JSON 数字会被解析为 float64）
		if f > 1e12 && f < 2e12 {
			// 可能是毫秒时间戳
			t := time.Unix(int64(f/1000), 0)
			return t.Format("2006-01-02 15:04:05")
		} else if f > 1e9 && f < 2e9 {
			// 可能是秒时间戳
			t := time.Unix(int64(f), 0)
			return t.Format("2006-01-02 15:04:05")
		}
		// 如果是整数，显示为整数
		if f == float64(int64(f)) {
			return fmt.Sprintf("%d", int64(f))
		}
		return fmt.Sprintf("%.2f", f)
	case reflect.Bool:
		if v.Bool() {
			return "✓"
		}
		return "✗"
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return "[]"
		}
		// 短数组直接显示
		if v.Len() <= 3 {
			b, _ := json.Marshal(v.Interface())
			s := string(b)
			if len(s) <= 50 {
				return s
			}
		}
		return fmt.Sprintf("[%d items]", v.Len())
	case reflect.Map:
		if v.Len() == 0 {
			return "{}"
		}
		return fmt.Sprintf("{%d keys}", v.Len())
	case reflect.Struct:
		// 尝试作为时间解析
		if t, ok := v.Interface().(time.Time); ok {
			return t.Format("2006-01-02 15:04:05")
		}
		b, _ := json.Marshal(v.Interface())
		s := string(b)
		if len(s) > 50 {
			return s[:47] + "..."
		}
		return s
	case reflect.Ptr:
		if v.IsNil() {
			return ""
		}
		return formatValue(v.Elem())
	default:
		s := fmt.Sprintf("%v", v.Interface())
		if len(s) > 50 {
			return s[:47] + "..."
		}
		return s
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
