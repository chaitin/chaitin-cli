package tanswer

import (
	"bytes"
	"encoding/json"
)

type SuccessEnvelope struct {
	Success  bool        `json:"success"`
	Task     string      `json:"task"`
	Command  string      `json:"command"`
	Query    any         `json:"query,omitempty"`
	Data     any         `json:"data,omitempty"`
	Warnings []string    `json:"warnings,omitempty"`
	Raw      interface{} `json:"raw,omitempty"`
}

type ErrorEnvelope struct {
	Success bool     `json:"success"`
	Task    string   `json:"task"`
	Command string   `json:"command"`
	Error   CLIError `json:"error"`
}

type CLIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Detail    any    `json:"detail,omitempty"`
}

func RenderJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
