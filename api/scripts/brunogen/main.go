// Command brunogen emits a Bruno (.bru) collection from openapi.yaml.
//
// One file per operation, grouped into folders by the first URL
// segment (files/, me/, study-guides/, etc). Path params surface as
// {{camelCaseName}} placeholders that resolve from the active
// environment; query params surface in a params:query block with
// their default value or an empty placeholder; JSON request bodies
// get an example synthesized from the schema.
//
// Run via `make bruno` or `go run ./scripts/brunogen`. The
// collection root (bruno/bruno.json, bruno/collection.bru,
// bruno/environments/) is hand-maintained and never touched by
// this tool; only the per-folder operation .bru files are written.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	specPath       = "openapi.yaml"
	collectionRoot = "bruno"
)

// methodOrder canonicalises the render order when an operation has
// multiple methods on the same path (rare but possible -- the
// detail routes on /study-guides/{id} have GET + PATCH + DELETE).
var methodOrder = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "brunogen:", err)
		os.Exit(1)
	}
}

func run() error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return fmt.Errorf("load %s: %w", specPath, err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		return fmt.Errorf("validate %s: %w", specPath, err)
	}

	// Wipe every per-folder tree under collectionRoot on each run so
	// stale .bru files from removed operations don't linger.
	// Environment files + bruno.json + collection.bru are untouched
	// (they live at the root, not inside operation folders).
	if err := cleanGeneratedFolders(collectionRoot); err != nil {
		return fmt.Errorf("clean: %w", err)
	}

	ops := collectOps(doc)
	if err := writeOps(ops); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	fmt.Printf("brunogen: wrote %d operations to %s/\n", len(ops), collectionRoot)
	return nil
}

// operation is the flattened view brunogen actually renders from.
type operation struct {
	Folder   string
	Seq      int
	Name     string
	FileName string
	Method   string
	Path     string
	Summary  string
	Desc     string
	Query    []param
	PathVars []string // camelCase names pre-substituted into URL
	BodyJSON string   // empty when no JSON body
	HasBody  bool
}

type param struct {
	Name     string // wire name (snake_case, matches spec)
	Default  string // first enum, default, or placeholder
	Disabled bool   // render with `~key:` prefix (no default -> disabled)
}

// collectOps walks the doc and builds an operation slice sorted
// first by folder, then by method order (GET, POST, ...), then by
// path. Sequence numbers are assigned per folder in that final
// order so Bruno's sidebar matches the render order deterministically.
func collectOps(doc *openapi3.T) []operation {
	var ops []operation
	paths := doc.Paths.Map()
	pathKeys := make([]string, 0, len(paths))
	for p := range paths {
		pathKeys = append(pathKeys, p)
	}
	sort.Strings(pathKeys)

	for _, p := range pathKeys {
		pi := paths[p]
		for _, m := range methodOrder {
			op := operationForMethod(pi, m)
			if op == nil {
				continue
			}
			ops = append(ops, buildOp(doc, p, m, op))
		}
	}

	sort.SliceStable(ops, func(i, j int) bool {
		if ops[i].Folder != ops[j].Folder {
			return ops[i].Folder < ops[j].Folder
		}
		mi := methodOrderIndex(ops[i].Method)
		mj := methodOrderIndex(ops[j].Method)
		if mi != mj {
			return mi < mj
		}
		return ops[i].Path < ops[j].Path
	})

	// Assign sequence numbers per folder.
	seq := map[string]int{}
	for i := range ops {
		seq[ops[i].Folder]++
		ops[i].Seq = seq[ops[i].Folder]
	}
	return ops
}

func operationForMethod(pi *openapi3.PathItem, method string) *openapi3.Operation {
	switch method {
	case "GET":
		return pi.Get
	case "POST":
		return pi.Post
	case "PUT":
		return pi.Put
	case "PATCH":
		return pi.Patch
	case "DELETE":
		return pi.Delete
	}
	return nil
}

func methodOrderIndex(m string) int {
	for i, v := range methodOrder {
		if v == m {
			return i
		}
	}
	return len(methodOrder)
}

// buildOp pulls the renderable fields out of the spec types.
func buildOp(_ *openapi3.T, path, method string, op *openapi3.Operation) operation {
	folder := folderFor(path)
	name := titleFromOperationID(op.OperationID)
	fileName := name + ".bru"

	out := operation{
		Folder:   folder,
		Name:     name,
		FileName: fileName,
		Method:   method,
		Path:     path,
		Summary:  op.Summary,
		Desc:     strings.TrimSpace(op.Description),
	}

	// Path-param vars: `{file_id}` -> `{{fileId}}` in the URL, and
	// the camelCase name gets added to PathVars for any per-request
	// docs that want to enumerate them.
	for _, pref := range op.Parameters {
		if pref == nil || pref.Value == nil {
			continue
		}
		p := pref.Value
		switch p.In {
		case openapi3.ParameterInPath:
			out.PathVars = append(out.PathVars, snakeToCamel(p.Name))
		case openapi3.ParameterInQuery:
			out.Query = append(out.Query, queryParam(p))
		}
	}

	// Request body: only JSON content is supported (the spec uses
	// application/json exclusively). Synthesized example comes from
	// schema.Example when provided, otherwise from a recursive
	// walk that emits placeholder values by type.
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		if m := op.RequestBody.Value.Content.Get("application/json"); m != nil {
			body := pickJSONExample(m)
			if body != "" {
				out.HasBody = true
				out.BodyJSON = body
			}
		}
	}

	return out
}

// folderFor groups operations into Bruno sidebar folders. The first
// non-variable segment is the group name -- /files/{id}/grants goes
// under "Files", /me/study-guides under "Me", etc. Folder names are
// Title-Cased for readability in Bruno.
func folderFor(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for _, p := range parts {
		if p == "" || strings.HasPrefix(p, "{") {
			continue
		}
		return toTitle(strings.ReplaceAll(p, "-", " "))
	}
	return "Misc"
}

// titleFromOperationID converts a PascalCase operationId ("ListFiles")
// into a spaced title ("List Files") suitable for the Bruno meta
// name + .bru filename.
func titleFromOperationID(opID string) string {
	if opID == "" {
		return "Unnamed"
	}
	var b strings.Builder
	for i, r := range opID {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// toTitle lowercases then upper-cases the first rune of each word.
// Used for folder names.
func toTitle(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if w == "" {
			continue
		}
		runes := []rune(strings.ToLower(w))
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// snakeToCamel converts snake_case to camelCase for Bruno var names.
// The spec uses snake_case path params (file_id); Bruno env vars are
// conventionally camelCase (fileId).
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		runes := []rune(parts[i])
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}

// queryParam extracts the Bruno render fields from an OpenAPI
// parameter. When the schema declares a default or an enum, use the
// first known-good value; otherwise emit the key as disabled so
// Bruno shows it in the UI with an empty value rather than sending
// a garbage query string on first open.
func queryParam(p *openapi3.Parameter) param {
	out := param{Name: p.Name, Disabled: true}
	if p.Schema == nil || p.Schema.Value == nil {
		return out
	}
	sv := p.Schema.Value
	if sv.Default != nil {
		out.Default = fmt.Sprintf("%v", sv.Default)
		out.Disabled = false
		return out
	}
	if len(sv.Enum) > 0 {
		out.Default = fmt.Sprintf("%v", sv.Enum[0])
		out.Disabled = false
		return out
	}
	// Otherwise stay disabled with an empty value -- user fills in.
	return out
}

// pickJSONExample chooses the best JSON body example for a
// requestBody: the explicit example on the MediaType, the schema's
// example, or a synthesized placeholder object walked from the
// schema. Returned string is a valid JSON document, ready to drop
// into a Bruno body:json block.
func pickJSONExample(m *openapi3.MediaType) string {
	if m.Example != nil {
		return prettyJSON(m.Example)
	}
	if m.Examples != nil {
		for _, ex := range m.Examples {
			if ex.Value != nil && ex.Value.Value != nil {
				return prettyJSON(ex.Value.Value)
			}
		}
	}
	if m.Schema != nil {
		return prettyJSON(synthesizeFromSchema(m.Schema, map[string]bool{}))
	}
	return ""
}

// synthesizeFromSchema recursively builds a Go value that mirrors
// the schema's structure, using the schema's Example when present
// and placeholder values otherwise. The visited map guards against
// self-referential schemas (rare but possible with $ref loops).
func synthesizeFromSchema(ref *openapi3.SchemaRef, visited map[string]bool) interface{} {
	if ref == nil || ref.Value == nil {
		return nil
	}
	if ref.Ref != "" {
		if visited[ref.Ref] {
			return nil
		}
		visited[ref.Ref] = true
		defer delete(visited, ref.Ref)
	}
	s := ref.Value
	if s.Example != nil {
		return s.Example
	}
	if s.Default != nil {
		return s.Default
	}
	if len(s.Enum) > 0 {
		return s.Enum[0]
	}

	// kin-openapi uses a Types slice; for schemas we care about,
	// exactly one type is declared.
	tp := ""
	if s.Type != nil && len(*s.Type) > 0 {
		tp = (*s.Type)[0]
	}
	switch tp {
	case "object":
		return synthesizeObject(s, visited)
	case "array":
		if s.Items == nil {
			return []interface{}{}
		}
		return []interface{}{synthesizeFromSchema(s.Items, visited)}
	case "string":
		return placeholderString(s.Format)
	case "integer", "number":
		return 0
	case "boolean":
		return false
	}
	// Composite schemas (allOf/oneOf/anyOf): walk the first variant.
	if len(s.AllOf) > 0 {
		return synthesizeFromSchema(s.AllOf[0], visited)
	}
	if len(s.OneOf) > 0 {
		return synthesizeFromSchema(s.OneOf[0], visited)
	}
	if len(s.AnyOf) > 0 {
		return synthesizeFromSchema(s.AnyOf[0], visited)
	}
	return nil
}

// synthesizeObject walks an object schema's Properties in
// deterministic key order so the generated JSON is stable across runs.
func synthesizeObject(s *openapi3.Schema, visited map[string]bool) map[string]interface{} {
	out := map[string]interface{}{}
	keys := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out[k] = synthesizeFromSchema(s.Properties[k], visited)
	}
	return out
}

// placeholderString picks a sensible string value for a given
// format hint so copy-paste users get a body that passes surface
// validation without hand-editing every field.
func placeholderString(format string) string {
	switch format {
	case "uuid":
		return "00000000-0000-0000-0000-000000000000"
	case "date-time":
		return "2026-04-20T00:00:00Z"
	case "date":
		return "2026-04-20"
	case "email":
		return "user@example.com"
	case "uri", "url":
		return "https://example.com"
	}
	return "string"
}

// prettyJSON formats a value with two-space indentation so the
// Bruno editor renders it readably. On error (should not happen for
// types we construct), fall back to a best-effort %#v.
func prettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%#v", v)
	}
	return string(b)
}

// cleanGeneratedFolders removes every subdirectory of root except
// `environments/` (hand-maintained). Files directly in root
// (bruno.json, collection.bru) are preserved.
func cleanGeneratedFolders(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(root, 0o755)
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "environments" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, e.Name())); err != nil {
			return fmt.Errorf("remove %s: %w", e.Name(), err)
		}
	}
	return nil
}

// writeOps creates the per-folder directories and writes one .bru
// file per operation.
func writeOps(ops []operation) error {
	for _, op := range ops {
		dir := filepath.Join(collectionRoot, op.Folder)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
		content := renderOp(op)
		path := filepath.Join(dir, op.FileName)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

// renderOp produces the full .bru text for one operation. The
// block order follows Bruno's canonical layout so Bruno doesn't
// rewrite files it reads -- saves noise in `make bruno` diffs.
func renderOp(op operation) string {
	var b strings.Builder

	// meta block
	fmt.Fprintf(&b, "meta {\n  name: %s\n  type: http\n  seq: %d\n}\n\n", op.Name, op.Seq)

	// method block
	url := substitutePathVars(op.Path)
	method := strings.ToLower(op.Method)
	bodyMode := "none"
	if op.HasBody {
		bodyMode = "json"
	}
	fmt.Fprintf(&b, "%s {\n  url: {{baseUrl}}/api%s\n  body: %s\n  auth: inherit\n}\n\n", method, url, bodyMode)

	// query params
	if len(op.Query) > 0 {
		b.WriteString("params:query {\n")
		for _, q := range op.Query {
			prefix := ""
			if q.Disabled {
				prefix = "~"
			}
			fmt.Fprintf(&b, "  %s%s: %s\n", prefix, q.Name, q.Default)
		}
		b.WriteString("}\n\n")
	}

	// body block
	if op.HasBody {
		b.WriteString("body:json {\n")
		// Indent the JSON by two spaces to nest inside the .bru block.
		for _, line := range strings.Split(op.BodyJSON, "\n") {
			if line == "" {
				b.WriteString("\n")
				continue
			}
			b.WriteString("  ")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("}\n\n")
	}

	// docs block
	if op.Summary != "" || op.Desc != "" {
		b.WriteString("docs {\n")
		if op.Summary != "" {
			fmt.Fprintf(&b, "  # %s\n\n", op.Summary)
		}
		if op.Desc != "" {
			for _, line := range strings.Split(op.Desc, "\n") {
				b.WriteString("  ")
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
		b.WriteString("}\n")
	}

	return b.String()
}

// substitutePathVars converts `/files/{file_id}/view` to
// `/files/{{fileId}}/view` so Bruno resolves the camelCase var from
// the active environment.
func substitutePathVars(path string) string {
	var b strings.Builder
	i := 0
	for i < len(path) {
		if path[i] != '{' {
			b.WriteByte(path[i])
			i++
			continue
		}
		j := strings.IndexByte(path[i:], '}')
		if j < 0 {
			// Malformed path -- write the rest verbatim.
			b.WriteString(path[i:])
			break
		}
		name := path[i+1 : i+j]
		b.WriteString("{{")
		b.WriteString(snakeToCamel(name))
		b.WriteString("}}")
		i += j + 1
	}
	return b.String()
}
