package cloudatlas

import "encoding/json"

type Config struct {
	URL     string `yaml:"url"`
	Token   string `yaml:"token"`
	SpaceID string `yaml:"space_id"`
}

type OpenAPI struct {
	OpenAPI    string              `yaml:"openapi"`
	Info       Info                `yaml:"info"`
	Paths      map[string]PathItem `yaml:"paths"`
	Components Components          `yaml:"components"`
	Security   []map[string]any    `yaml:"security"`
}

type Info struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

type Components struct {
	Schemas         map[string]Schema         `yaml:"schemas"`
	SecuritySchemes map[string]SecurityScheme `yaml:"securitySchemes"`
}

type SecurityScheme struct {
	Type string `yaml:"type"`
	In   string `yaml:"in"`
	Name string `yaml:"name"`
}

type PathItem struct {
	Get    *Operation `yaml:"get,omitempty"`
	Post   *Operation `yaml:"post,omitempty"`
	Put    *Operation `yaml:"put,omitempty"`
	Delete *Operation `yaml:"delete,omitempty"`
	Patch  *Operation `yaml:"patch,omitempty"`
}

type Operation struct {
	OperationID string       `yaml:"operationId"`
	Summary     string       `yaml:"summary"`
	Description string       `yaml:"description"`
	Tags        []string     `yaml:"tags"`
	Parameters  []Parameter  `yaml:"parameters"`
	RequestBody *RequestBody `yaml:"requestBody"`
}

type Parameter struct {
	Name        string  `yaml:"name"`
	In          string  `yaml:"in"`
	Description string  `yaml:"description"`
	Required    bool    `yaml:"required"`
	Style       string  `yaml:"style"`
	Explode     *bool   `yaml:"explode"`
	Example     any     `yaml:"example"`
	Schema      *Schema `yaml:"schema"`
}

type RequestBody struct {
	Content map[string]MediaType `yaml:"content"`
}

type MediaType struct {
	Schema  *Schema `yaml:"schema"`
	Example any     `yaml:"example"`
}

type Schema struct {
	Type        string            `yaml:"type" json:"type,omitempty"`
	Title       string            `yaml:"title" json:"title,omitempty"`
	Description string            `yaml:"description" json:"description,omitempty"`
	Ref         string            `yaml:"$ref" json:"$ref,omitempty"`
	Items       *Schema           `yaml:"items" json:"items,omitempty"`
	Properties  map[string]Schema `yaml:"properties" json:"properties,omitempty"`
	Required    []string          `yaml:"required" json:"required,omitempty"`
	Enum        []any             `yaml:"enum" json:"enum,omitempty"`
	Default     any               `yaml:"default" json:"default,omitempty"`
	Example     any               `yaml:"example" json:"example,omitempty"`
	Nullable    bool              `yaml:"nullable" json:"nullable,omitempty"`
	XApifox     ApifoxSchema      `yaml:"x-apifox" json:"x-apifox,omitempty"`
	XApifoxEnum []ApifoxEnumItem  `yaml:"x-apifox-enum" json:"x-apifox-enum,omitempty"`
}

type ApifoxSchema struct {
	EnumDescriptions map[string]string `yaml:"enumDescriptions" json:"enumDescriptions,omitempty"`
}

type ApifoxEnumItem struct {
	Value       any    `yaml:"value" json:"value,omitempty"`
	Name        string `yaml:"name" json:"name,omitempty"`
	Description string `yaml:"description" json:"description,omitempty"`
}

type operationRef struct {
	method string
	path   string
	op     *Operation
}

type APIEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type ListEnvelope struct {
	Current int             `json:"current"`
	Size    int             `json:"size"`
	Total   int             `json:"total"`
	Items   json.RawMessage `json:"items"`
}
