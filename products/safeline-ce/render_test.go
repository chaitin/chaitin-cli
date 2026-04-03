package safelinece

import (
	"bytes"
	"testing"
)

func TestJSONRenderer_Render(t *testing.T) {
	var buf bytes.Buffer
	renderer := &JSONRenderer{out: &buf}

	data := map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": 1, "name": "test"},
		},
	}

	err := renderer.Render(data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	expected := `{
  "data": [
    {
      "id": 1,
      "name": "test"
    }
  ]
}
`
	if buf.String() != expected {
		t.Errorf("Render() = %q, want %q", buf.String(), expected)
	}
}

func TestTableRenderer_Render(t *testing.T) {
	t.Run("slice data", func(t *testing.T) {
		var buf bytes.Buffer
		renderer := &TableRenderer{out: &buf}

		data := map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": float64(1), "name": "test1"},
				{"id": float64(2), "name": "test2"},
			},
		}

		err := renderer.Render(data)
		if err != nil {
			t.Fatalf("Render() error = %v", err)
		}

		output := buf.String()
		if !containsAll(output, "ID", "NAME", "1", "test1", "2", "test2") {
			t.Errorf("Render() output missing expected columns: %s", output)
		}
	})

	t.Run("single object", func(t *testing.T) {
		var buf bytes.Buffer
		renderer := &TableRenderer{out: &buf}

		data := map[string]interface{}{
			"data": map[string]interface{}{"id": float64(1), "name": "test"},
		}

		err := renderer.Render(data)
		if err != nil {
			t.Fatalf("Render() error = %v", err)
		}

		output := buf.String()
		if !containsAll(output, "ID", "NAME", "1", "test") {
			t.Errorf("Render() output missing expected columns: %s", output)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		var buf bytes.Buffer
		renderer := &TableRenderer{out: &buf}

		data := map[string]interface{}{
			"data": []interface{}{},
		}

		err := renderer.Render(data)
		if err != nil {
			t.Fatalf("Render() error = %v", err)
		}

		if buf.String() != "No data found\n" {
			t.Errorf("Render() = %q, want %q", buf.String(), "No data found\n")
		}
	})

	t.Run("nil data", func(t *testing.T) {
		var buf bytes.Buffer
		renderer := &TableRenderer{out: &buf}

		err := renderer.Render(nil)
		if err != nil {
			t.Fatalf("Render() error = %v", err)
		}

		if buf.String() != "No data found\n" {
			t.Errorf("Render() = %q, want %q", buf.String(), "No data found\n")
		}
	})
}

func TestNewRenderer(t *testing.T) {
	var buf bytes.Buffer

	jsonRenderer := NewRenderer(FormatJSON, &buf)
	if _, ok := jsonRenderer.(*JSONRenderer); !ok {
		t.Error("NewRenderer(FormatJSON) should return *JSONRenderer")
	}

	tableRenderer := NewRenderer(FormatTable, &buf)
	if _, ok := tableRenderer.(*TableRenderer); !ok {
		t.Error("NewRenderer(FormatTable) should return *TableRenderer")
	}
}

func TestFormatColumnName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"id", "ID"},
		{"name", "NAME"},
		{"created_at", "CREATED_AT"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := formatColumnName(tt.input); got != tt.expected {
				t.Errorf("formatColumnName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !bytes.Contains([]byte(s), []byte(substr)) {
			return false
		}
	}
	return true
}
