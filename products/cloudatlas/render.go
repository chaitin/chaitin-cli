package cloudatlas

import (
	"encoding/json"
	"fmt"
	"io"
)

// Renderer prints an API response. JSON is the only supported format: every
// command emits the response body as indented JSON, so list responses keep
// their paging envelope (current/size/total/items).
type Renderer struct {
	out io.Writer
}

func NewRenderer(format string, out io.Writer) (Renderer, error) {
	if format != "json" {
		return Renderer{}, fmt.Errorf("unsupported output format %q; only \"json\" is supported", format)
	}
	return Renderer{out: out}, nil
}

func (r Renderer) Render(data json.RawMessage) error {
	if len(data) == 0 || string(data) == "null" {
		_, err := fmt.Fprintln(r.out, "null")
		return err
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("parse response for render: %w", err)
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("render response: %w", err)
	}
	_, err = fmt.Fprintln(r.out, string(pretty))
	return err
}
