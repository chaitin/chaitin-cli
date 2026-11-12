package client

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

type Renderer struct {
	format Format
	out    io.Writer
}

func NewRenderer(format string, out io.Writer) Renderer {
	if format == string(FormatTable) {
		return Renderer{format: FormatTable, out: out}
	}
	return Renderer{format: FormatJSON, out: out}
}

// Render helps to output data into the console.
// By default, it parses standard Insight JSON response payloads
// and attempts to draw a table if it detects a list.
func (r Renderer) Render(rawData []byte) error {
	if len(rawData) == 0 {
		return nil
	}

	if r.format == FormatJSON {
		// Output raw JSON as it is
		_, err := r.out.Write(rawData)
		if err == nil {
			_, err = r.out.Write([]byte("\n"))
		}
		return err
	}

	// Try to unpack response structure
	// Insight endpoints often wrap arrays in an object, e.g. {"data": [...]} or returning an array directly
	var result any
	if err := json.Unmarshal(rawData, &result); err != nil {
		// Fallback to plain string output if not JSON
		_, err = fmt.Fprintln(r.out, string(rawData))
		return err
	}

	// For simple extraction of lists, if "data" is an array, we extract it.
	var items []any
	if obj, ok := result.(map[string]any); ok {
		if dataField, hasData := obj["data"]; hasData {
			if arr, ok := dataField.([]any); ok {
				items = arr
			} else {
				// If data is just a single object, render it pretty
				return r.renderPretty(result)
			}
		} else {
			return r.renderPretty(result)
		}
	} else if arr, ok := result.([]any); ok {
		items = arr
	} else {
		return r.renderPretty(result)
	}

	if len(items) == 0 {
		fmt.Fprintln(r.out, "No data available.")
		return nil
	}

	return r.renderTable(items)
}

func (r Renderer) renderPretty(value any) error {
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(r.out, string(pretty))
	return err
}

func (r Renderer) renderTable(items []any) error {
	if len(items) == 0 {
		return nil
	}

	// Use standard text/tabwriter
	w := tabwriter.NewWriter(r.out, 0, 0, 3, ' ', 0)

	// Extract headers from the first item if it's a map
	firstItem, ok := items[0].(map[string]any)
	if !ok {
		// Cannot tableify non-objects easily, fallback
		return r.renderPretty(items)
	}

	var headers []string
	for k := range firstItem {
		headers = append(headers, k)
	}

	// Print headers
	for _, h := range headers {
		fmt.Fprintf(w, "%s\t", h)
	}
	fmt.Fprintln(w)

	// Print rows
	for _, itemAny := range items {
		itemMap, ok := itemAny.(map[string]any)
		if !ok {
			continue
		}
		for _, h := range headers {
			val := itemMap[h]
			// format value nicely
			fmt.Fprintf(w, "%v\t", formatValue(val))
		}
		fmt.Fprintln(w)
	}

	return w.Flush()
}

func formatValue(val any) string {
	if val == nil {
		return "-"
	}
	switch v := val.(type) {
	case string:
		// Truncate long strings
		if len(v) > 60 {
			return v[:57] + "..."
		}
		return v
	case map[string]any, []any:
		b, _ := json.Marshal(v)
		if len(b) > 40 {
			return string(b[:37]) + "..."
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}

var _ = os.Stdout
