package main

/*
One-shot CLI command generator: reads openapi.json and writes concrete
cobra command Go source files into the products/safeline/modules/ directory.

Usage:
    go run cmd/gen-cli/main.go

Output (committed to repo, not generated at runtime):
    products/safeline/modules/<module>/<module>.go
    products/safeline/modules/register.go
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

func main() {
	specPath := filepath.Join("products", "safeline", "openapi.json")
	outputDir := filepath.Join("products", "safeline", "modules")

	data, err := os.ReadFile(specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", specPath, err)
		os.Exit(1)
	}

	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing spec: %v\n", err)
		os.Exit(1)
	}

	modules := groupByModule(spec)
	fmt.Fprintf(os.Stderr, "Generating commands for %d modules\n", len(modules))

	for name, ops := range modules {
		sort.Slice(ops, func(i, j int) bool { return ops[i].CmdName < ops[j].CmdName })
		if err := writeModuleFile(outputDir, name, ops); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", name, err)
		} else {
			fmt.Fprintf(os.Stderr, "  %s: %d operations\n", name, len(ops))
		}
	}

	if err := writeRegisterFile(outputDir, modules); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing register.go: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Done.\n")
}

// --- Types ---

type Spec struct {
	Paths map[string]PathItem `json:"paths"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
}

type Operation struct {
	OperationID string   `json:"operationId"`
	Tags        []string `json:"tags"`
	Summary     string   `json:"summary"`
	Parameters  []Param  `json:"parameters,omitempty"`
	RequestBody *ReqBody `json:"requestBody,omitempty"`
}

type Param struct {
	Name string `json:"name"`
	In   string `json:"in"`
}

type ReqBody struct {
	Content map[string]struct{} `json:"content"`
}

type ModuleOp struct {
	Path        string
	Method      string
	CmdName     string
	CmdVar      string
	HasBody     bool
	QueryParams []string
}

// --- Logic ---

func groupByModule(spec Spec) map[string][]ModuleOp {
	modules := make(map[string][]ModuleOp)
	seen := make(map[string]bool)

	for path, item := range spec.Paths {
		for _, op := range []*Operation{item.Get, item.Post, item.Put, item.Delete, item.Patch} {
			if op == nil || len(op.Tags) == 0 {
				continue
			}
			tag := op.Tags[0]
			method := methodOf(item, op)
			cmdName := toCmdName(op.Summary, op.OperationID)
			if cmdName == "" {
				continue
			}

			key := tag + ":" + method + ":" + path
			if seen[key] {
				continue
			}
			seen[key] = true

			mop := ModuleOp{
				Path:    path,
				Method:  method,
				CmdName: toCmdName(op.Summary, op.OperationID) + "-" + strings.ToLower(method),
				HasBody: op.RequestBody != nil,
			}
			for _, p := range op.Parameters {
				if p.In == "query" && p.Name != "count" && p.Name != "offset" && p.Name != "server_ts" {
					mop.QueryParams = append(mop.QueryParams, p.Name)
				}
			}
			modules[tag] = append(modules[tag], mop)
		}
	}
	return modules
}

func methodOf(item PathItem, op *Operation) string {
	switch {
	case item.Get == op:
		return "GET"
	case item.Post == op:
		return "POST"
	case item.Put == op:
		return "PUT"
	case item.Delete == op:
		return "DELETE"
	case item.Patch == op:
		return "PATCH"
	}
	return "GET"
}

func toCmdName(summary, opID string) string {
	name := summary
	if name == "" {
		name = opID
	}
	name = strings.TrimSuffix(name, "API")
	name = strings.TrimSuffix(name, "Api")

	var result []rune
	for i, r := range name {
		if r >= 'A' && r <= 'Z' && i > 0 {
			result = append(result, '-')
		}
		result = append(result, []rune(strings.ToLower(string(r)))...)
	}
	s := strings.Trim(string(result), "-")
	if s == "" {
		return "unknown"
	}
	return s
}

// --- Templates ---

var moduleTmpl = template.Must(template.New("module").Parse(`// Code generated from openapi.json. DO NOT EDIT.
package {{.Name}}

import (
	"fmt"
	"io"
	"os"

	"github.com/chaitin/workspace-cli/products/safeline/pkg/client"
	"github.com/spf13/cobra"
)

func NewCmd(clFn func() *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "{{.Name}}",
		Short: "{{.Name}} API commands",
	}
{{- range .Ops}}
	cmd.AddCommand({{.CmdVar}}Cmd(clFn))
{{- end}}
	return cmd
}
{{range $i, $op := .Ops}}

var {{.CmdVar}}Path = "{{$op.Path}}"

func {{.CmdVar}}Cmd(clFn func() *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "{{$op.CmdName}}",
		Short: "{{$op.Method}} {{$op.Path}}",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clFn()
	{{- if $op.HasBody}}
			input, err := readInput(cmd)
			if err != nil {
				return err
			}
			env, err := c.Do("{{$op.Method}}", {{$op.CmdVar}}Path, input, buildQuery(cmd))
	{{- else}}
			env, err := c.Do("{{$op.Method}}", {{$op.CmdVar}}Path, nil, buildQuery(cmd))
	{{- end}}
			if err != nil {
				return err
			}
			return printEnvelope(cmd, env)
		},
	}
	{{- if $op.HasBody}}
	cmd.Flags().String("file", "", "JSON input file (default: stdin)")
	{{- end}}
	{{- range $op.QueryParams}}
	cmd.Flags().String("{{.}}", "", "{{.}} filter")
	{{- end}}
	return cmd
}
{{- end}}

func buildQuery(cmd *cobra.Command) map[string]string {
	q := make(map[string]string)
{{- range .Ops}}
	{{- range .QueryParams}}
	if v, _ := cmd.Flags().GetString("{{.}}"); v != "" {
		q["{{.}}"] = v
	}
	{{- end}}
{{- end}}
	if v, _ := cmd.Flags().GetString("count"); v != "" {
		q["count"] = v
	}
	if v, _ := cmd.Flags().GetString("offset"); v != "" {
		q["offset"] = v
	}
	return q
}

func readInput(cmd *cobra.Command) (io.Reader, error) {
	file, _ := cmd.Flags().GetString("file")
	if file == "" || file == "-" {
		return os.Stdin, nil
	}
	return os.Open(file)
}

func printEnvelope(cmd *cobra.Command, env *client.Envelope) error {
	if env.Msg != nil && env.Msg.Level == "warning" {
		fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: %s\n", env.Msg.Text)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(env.Data))
	return nil
}

var _ = fmt.Sprintf
`))

func writeModuleFile(outputDir, name string, ops []ModuleOp) error {
	dir := filepath.Join(outputDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Ensure unique var names using index
	for i := range ops {
		ops[i].CmdVar = fmt.Sprintf("%s_%s_%d", name, strings.ReplaceAll(ops[i].CmdName, "-", "_"), i)
	}

	f, err := os.Create(filepath.Join(dir, name+".go"))
	if err != nil {
		return err
	}
	defer f.Close()

	return moduleTmpl.Execute(f, struct {
		Name string
		Ops  []ModuleOp
	}{Name: name, Ops: ops})
}

var registerTmpl = template.Must(template.New("register").Parse(`// Code generated from openapi.json. DO NOT EDIT.
package modules

import (
	"github.com/chaitin/workspace-cli/products/safeline/pkg/client"
	"github.com/spf13/cobra"
)
{{range $name := .}}
import {{$name}} "github.com/chaitin/workspace-cli/products/safeline/modules/{{$name}}"
{{end}}

// RegisterAll registers all safeline module commands.
func RegisterAll(root *cobra.Command, clFn func() *client.Client) {
{{range $name := .}}
	root.AddCommand({{$name}}.NewCmd(clFn))
{{end}}
}
`))

func writeRegisterFile(outputDir string, modules map[string][]ModuleOp) error {
	f, err := os.Create(filepath.Join(outputDir, "register.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	var names []string
	for name := range modules {
		names = append(names, name)
	}
	sort.Strings(names)
	return registerTmpl.Execute(f, names)
}
