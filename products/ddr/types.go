package ddr

import "encoding/json"

const (
	ExitSuccess      = 0
	ExitAPIError     = 1
	ExitNetworkError = 2
	ExitConfigError  = 3
)

type Config struct {
	URL       string `yaml:"url"`
	APIKey    string `yaml:"api_key"`
	CompanyID string `yaml:"company_id"`
}

type APIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type OpenAPI struct {
	OpenAPI string              `json:"openapi"`
	Info    Info                `json:"info"`
	Tags    []Tag               `json:"tags"`
	Paths   map[string]PathItem `json:"paths"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
}

type Operation struct {
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags"`
	Parameters  []Parameter  `json:"parameters"`
	RequestBody *RequestBody `json:"requestBody,omitempty"`
}

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description"`
	Required    bool    `json:"required"`
	Schema      *Schema `json:"schema,omitempty"`
}

type Schema struct {
	Type string `json:"type,omitempty"`
	Ref  string `json:"$ref,omitempty"`
}

type RequestBody struct {
	Required bool                 `json:"required"`
	Content  map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema  *Schema     `json:"schema,omitempty"`
	Example interface{} `json:"example,omitempty"`
}
