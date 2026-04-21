package ddr

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var versionSegmentPattern = regexp.MustCompile(`^v[0-9]+([._-][0-9]+)*$`)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) GenerateCommands(api *OpenAPI) ([]*cobra.Command, error) {
	tagDescriptions := buildTagDescriptions(api.Tags)
	parentCommands := make(map[string]*cobra.Command)
	keys := make([]string, 0, len(api.Paths))
	for path := range api.Paths {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	for _, path := range keys {
		pathItem := api.Paths[path]
		operations := []struct {
			method string
			op     *Operation
		}{
			{method: "GET", op: pathItem.Get},
			{method: "POST", op: pathItem.Post},
			{method: "PUT", op: pathItem.Put},
			{method: "DELETE", op: pathItem.Delete},
			{method: "PATCH", op: pathItem.Patch},
		}

		for _, item := range operations {
			if item.op == nil {
				continue
			}

			parentName, childName, _ := classifyPath(path, item.method)
			parentCmd, ok := parentCommands[parentName]
			if !ok {
				parentCmd = &cobra.Command{
					Use:   parentName,
					Short: descriptionForCommand(tagDescriptions, parentName, parentName),
				}
				parentCommands[parentName] = parentCmd
			}

			target := parentCmd
			if childName != "" {
				target = getOrCreateChildCommand(parentCmd, childName, tagDescriptions)
			}
			target.AddCommand(p.createOperationCommand(item.method, path, item.op))
		}
	}

	var commands []*cobra.Command
	for _, name := range sortedKeys(parentCommands) {
		commands = append(commands, parentCommands[name])
	}
	return commands, nil
}

func sortedKeys(m map[string]*cobra.Command) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func buildTagDescriptions(tags []Tag) map[string]string {
	descriptions := make(map[string]string, len(tags))
	for _, tag := range tags {
		name := normalizeCommandSegment(tag.Name)
		if name == "" || strings.TrimSpace(tag.Description) == "" {
			continue
		}
		descriptions[name] = strings.TrimSpace(tag.Description)
	}
	return descriptions
}

func descriptionForCommand(descriptions map[string]string, fallbackKeys ...string) string {
	for _, key := range fallbackKeys {
		if description, ok := descriptions[normalizeCommandSegment(key)]; ok && description != "" {
			return description
		}
	}
	if len(fallbackKeys) > 0 {
		return fmt.Sprintf("%s commands", fallbackKeys[len(fallbackKeys)-1])
	}
	return "commands"
}

func getOrCreateChildCommand(parent *cobra.Command, child string, tagDescriptions map[string]string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Use == child {
			return cmd
		}
	}
	childCmd := &cobra.Command{
		Use:   child,
		Short: descriptionForCommand(tagDescriptions, parent.Use+"/"+child, child),
	}
	parent.AddCommand(childCmd)
	return childCmd
}

func classifyPath(path, method string) (string, string, string) {
	segments := normalizedSegments(path)
	if len(segments) == 0 {
		return "default", "", defaultOperationName(method)
	}

	parent := normalizeCommandSegment(segments[0])
	child := ""
	opStart := 1
	if len(segments) > 2 && !isParameterSegment(segments[1]) && !looksLikeLeafOperation(segments[1]) {
		child = normalizeCommandSegment(segments[1])
		opStart = 2
	}

	opName := operationName(segments[opStart:], method)
	return parent, child, opName
}

func normalizedSegments(path string) []string {
	raw := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, 0, len(raw))
	for _, segment := range raw {
		if segment == "" || versionSegmentPattern.MatchString(segment) {
			continue
		}
		segments = append(segments, segment)
	}
	return segments
}

func looksLikeLeafOperation(segment string) bool {
	switch strings.ToLower(segment) {
	case "list", "detail", "download", "stafflist", "timelinelist", "retrieve", "modify":
		return true
	}
	return strings.HasPrefix(segment, "{") || strings.HasSuffix(segment, "list")
}

func isParameterSegment(segment string) bool {
	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
}

func operationName(segments []string, method string) string {
	filtered := make([]string, 0, len(segments))
	for _, segment := range segments {
		if isParameterSegment(segment) {
			continue
		}
		filtered = append(filtered, segment)
	}

	if len(filtered) == 0 {
		return defaultOperationName(method)
	}

	if len(filtered) >= 2 && filtered[0] == "action" {
		return normalizeCommandSegment(strings.Join(filtered[1:], "-"))
	}

	last := filtered[len(filtered)-1]
	switch strings.ToLower(last) {
	case "detail":
		return "get"
	case "retrieve":
		return "get"
	case "modify":
		return "update"
	}

	return normalizeCommandSegment(strings.Join(filtered, "-"))
}

func defaultOperationName(method string) string {
	switch method {
	case "GET":
		return "get"
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	case "PATCH":
		return "patch"
	default:
		return strings.ToLower(method)
	}
}

func normalizeCommandSegment(segment string) string {
	segment = strings.Trim(segment, "/")
	segment = strings.Trim(segment, "{}")
	segment = strings.ReplaceAll(segment, "_", "-")
	segment = strings.ReplaceAll(segment, ".", "-")
	return strings.ToLower(segment)
}

func (p *Parser) createOperationCommand(method, path string, op *Operation) *cobra.Command {
	use := classifyOperationName(path, method)
	short := strings.TrimSpace(op.Summary)
	if short == "" {
		short = strings.TrimSpace(op.Description)
	}
	if short == "" {
		short = fmt.Sprintf("%s %s", method, path)
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.executeCommand(cmd, method, path, op)
		},
	}

	for _, param := range op.Parameters {
		if shouldSkipManagedHeader(param) {
			continue
		}
		addParamFlag(cmd, param)
	}

	if op.RequestBody != nil {
		cmd.Flags().String("body", "", "Request body as JSON string")
		cmd.Flags().String("body-file", "", "Path to JSON file used as request body")
	}

	return cmd
}

func classifyOperationName(path, method string) string {
	_, _, opName := classifyPath(path, method)
	return opName
}

func shouldSkipManagedHeader(param Parameter) bool {
	if strings.ToLower(param.In) != "header" {
		return false
	}
	name := strings.ToLower(param.Name)
	return name == "authorization" ||
		name == "x-cs-header-company" ||
		name == "x-cs-header-app" ||
		name == "x-cs-header-debug" ||
		name == "x-cs-header-crypt" ||
		name == "x-cs-header-timezone" ||
		name == "accept" ||
		name == "accept-language" ||
		name == "content-type" ||
		name == "if-none-match"
}

func addParamFlag(cmd *cobra.Command, param Parameter) {
	flagName := flagNameForParam(param)
	if cmd.Flags().Lookup(flagName) != nil {
		return
	}
	desc := param.Description
	if desc == "" {
		desc = fmt.Sprintf("%s parameter", param.Name)
	}

	switch schemaType(param.Schema) {
	case "integer":
		cmd.Flags().Int(flagName, 0, desc)
	case "boolean":
		cmd.Flags().Bool(flagName, false, desc)
	default:
		cmd.Flags().String(flagName, defaultStringValue(param), desc)
	}

	if param.Required && !shouldRelaxRequired(param) {
		_ = cmd.MarkFlagRequired(flagName)
	}
}

func defaultStringValue(param Parameter) string {
	switch {
	case strings.EqualFold(param.Name, "x-cs-header-app"):
		return "qzh"
	case strings.EqualFold(param.Name, "accept"):
		return "application/json"
	case strings.EqualFold(param.Name, "accept-language"):
		return "zh"
	case strings.EqualFold(param.Name, "content-type"):
		return "application/json"
	case strings.EqualFold(param.Name, "if-none-match"):
		return ""
	}
	return ""
}

func shouldRelaxRequired(param Parameter) bool {
	return strings.EqualFold(param.Name, "x-cs-header-debug") ||
		strings.EqualFold(param.Name, "x-cs-header-app") ||
		strings.EqualFold(param.Name, "x-cs-header-crypt") ||
		strings.EqualFold(param.Name, "x-cs-header-timezone") ||
		strings.EqualFold(param.Name, "accept") ||
		strings.EqualFold(param.Name, "accept-language") ||
		strings.EqualFold(param.Name, "content-type") ||
		strings.EqualFold(param.Name, "if-none-match")
}

func schemaType(schema *Schema) string {
	if schema == nil {
		return "string"
	}
	if schema.Type != "" {
		return schema.Type
	}
	return "string"
}

func flagNameForParam(param Parameter) string {
	name := strings.ToLower(param.Name)
	replacer := strings.NewReplacer("_", "-", ".", "-", " ", "-", "/", "-", "{", "", "}", "")
	return replacer.Replace(name)
}

func (p *Parser) executeCommand(cmd *cobra.Command, method, path string, op *Operation) error {
	query := make(map[string]string)
	headers := make(map[string]string)
	body, err := collectRequestBody(cmd, op.RequestBody)
	if err != nil {
		return err
	}

	apiPath := path
	for _, param := range op.Parameters {
		if shouldSkipManagedHeader(param) {
			continue
		}

		value, present, err := readFlagValue(cmd, param)
		if err != nil {
			return err
		}
		if !present {
			continue
		}

		switch strings.ToLower(param.In) {
		case "path":
			apiPath = strings.ReplaceAll(apiPath, "{"+param.Name+"}", value)
		case "query":
			query[param.Name] = value
		case "header":
			headers[param.Name] = value
		}
	}

	client := getClient(cmd)
	var result interface{}
	if err := client.Do(cmd.Context(), method, apiPath, mapToValues(query), headers, body, &result); err != nil {
		return err
	}

	return getRenderer(cmd).Render(result)
}

func readFlagValue(cmd *cobra.Command, param Parameter) (string, bool, error) {
	flag := cmd.Flags().Lookup(flagNameForParam(param))
	if flag == nil {
		return "", false, nil
	}

	if !flag.Changed {
		if fallback := defaultStringValue(param); fallback != "" {
			return fallback, true, nil
		}
		return "", false, nil
	}

	switch schemaType(param.Schema) {
	case "integer":
		value, err := cmd.Flags().GetInt(flag.Name)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("%d", value), true, nil
	case "boolean":
		value, err := cmd.Flags().GetBool(flag.Name)
		if err != nil {
			return "", false, err
		}
		if value {
			return "true", true, nil
		}
		return "false", true, nil
	default:
		value, err := cmd.Flags().GetString(flag.Name)
		if err != nil {
			return "", false, err
		}
		return value, true, nil
	}
}

func collectRequestBody(cmd *cobra.Command, requestBody *RequestBody) (interface{}, error) {
	if requestBody == nil {
		return nil, nil
	}

	bodyFile, err := cmd.Flags().GetString("body-file")
	if err != nil {
		return nil, err
	}
	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("read body file %s: %w", bodyFile, err)
		}
		return parseJSONBody(data)
	}

	bodyText, err := cmd.Flags().GetString("body")
	if err != nil {
		return nil, err
	}
	if bodyText == "" {
		return nil, nil
	}

	return parseJSONBody([]byte(bodyText))
}

func parseJSONBody(data []byte) (interface{}, error) {
	var body interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("parse request body JSON: %w", err)
	}
	return body, nil
}

func mapToValues(values map[string]string) url.Values {
	result := make(url.Values, len(values))
	for key, value := range values {
		result[key] = []string{value}
	}
	return result
}
