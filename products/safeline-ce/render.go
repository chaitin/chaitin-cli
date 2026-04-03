package safelinece

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
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

func (r *TableRenderer) renderSlice(val reflect.Value) error {
	if val.Len() == 0 {
		fmt.Fprintln(r.out, "No data found")
		return nil
	}

	// 获取列名
	columns := inferColumns(val.Index(0).Interface())
	if len(columns) == 0 {
		fmt.Fprintln(r.out, "No data found")
		return nil
	}

	// 打印表头
	r.printRow(columns)

	// 打印分隔线
	separators := make([]string, len(columns))
	for i, col := range columns {
		separators[i] = strings.Repeat("-", len(col))
	}
	r.printRow(separators)

	// 打印数据行
	for i := 0; i < val.Len(); i++ {
		row := extractRow(val.Index(i).Interface(), columns)
		r.printRow(row)
	}

	return nil
}

func (r *TableRenderer) renderSingle(val reflect.Value) error {
	columns := inferColumns(val.Interface())
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
	r.printRow(extractRow(val.Interface(), columns))

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

func extractData(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	// 尝试从 map 中提取 data 字段
	if m, ok := data.(map[string]interface{}); ok {
		if d, exists := m["data"]; exists {
			return d
		}
		return m
	}

	return data
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
			columns = append(columns, formatColumnName(key.String()))
		}
	} else {
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				name := strings.Split(jsonTag, ",")[0]
				if name != "" {
					columns = append(columns, formatColumnName(name))
				}
			} else {
				columns = append(columns, formatColumnName(field.Name))
			}
		}
	}

	sort.Strings(columns)
	return columns
}

func formatColumnName(name string) string {
	return strings.ToUpper(name)
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
				return formatValue(v)
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
				return formatValue(val.Field(i))
			}
		}
	}

	return ""
}

func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct:
		b, _ := json.Marshal(v.Interface())
		return string(b)
	case reflect.Ptr:
		if v.IsNil() {
			return ""
		}
		return formatValue(v.Elem())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
