package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Operation is a resolved OpenAPI operation ready for tool translation.
type Operation struct {
	ID          string
	Method      string
	Path        string
	Summary     string
	Description string
	RequestBody json.RawMessage // JSON Schema for the request body, if any.
	Parameters  []Parameter
}

// Parameter is a path or query parameter from an OpenAPI operation.
type Parameter struct {
	Name        string          `json:"name"`
	In          string          `json:"in"` // "path" or "query"
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	Required    bool            `json:"required,omitempty"`
}

// ToolFromOperation converts an OpenAPI operation into an MCP tool definition.
func ToolFromOperation(op Operation) Tool {
	description := op.Summary
	if op.Description != "" {
		if description != "" {
			description += " — "
		}
		description += op.Description
	}
	if description == "" {
		description = fmt.Sprintf("%s %s", op.Method, op.Path)
	}

	schema := buildInputSchema(op)

	return Tool{
		Name:        toolName(op),
		Description: description,
		InputSchema: schema,
	}
}

// OperationsFromSpec extracts operations from a parsed OpenAPI 3.x spec.
// The spec is expected to be a JSON-decoded map.
func OperationsFromSpec(spec map[string]any) []Operation {
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		return nil
	}

	var ops []Operation
	for path, pathItem := range paths {
		methods, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		for method, opAny := range methods {
			method = strings.ToUpper(method)
			if !isHTTPMethod(method) {
				continue
			}
			opMap, ok := opAny.(map[string]any)
			if !ok {
				continue
			}

			op := Operation{
				Method: method,
				Path:   path,
			}
			if v, ok := opMap["operationId"].(string); ok {
				op.ID = v
			}
			if v, ok := opMap["summary"].(string); ok {
				op.Summary = v
			}
			if v, ok := opMap["description"].(string); ok {
				op.Description = v
			}
			op.Parameters = extractParameters(opMap)
			op.RequestBody = extractRequestBodySchema(opMap)

			ops = append(ops, op)
		}
	}
	return ops
}

func toolName(op Operation) string {
	if op.ID != "" {
		return sanitizeName(op.ID)
	}
	name := strings.ToLower(op.Method) + "_" + strings.ReplaceAll(strings.Trim(op.Path, "/"), "/", "_")
	return sanitizeName(name)
}

func sanitizeName(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_', c == '-', c == '.':
			b.WriteRune(c)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func isHTTPMethod(m string) bool {
	switch m {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	}
	return false
}

func buildInputSchema(op Operation) json.RawMessage {
	properties := map[string]json.RawMessage{}
	var required []string

	for _, p := range op.Parameters {
		prop := p.Schema
		if prop == nil {
			prop = json.RawMessage(`{"type":"string"}`)
		}
		if p.Description != "" {
			// Merge description into schema.
			var m map[string]any
			if err := json.Unmarshal(prop, &m); err == nil {
				m["description"] = p.Description
				if merged, err := json.Marshal(m); err == nil {
					prop = merged
				}
			}
		}
		properties[p.Name] = prop
		if p.Required {
			required = append(required, p.Name)
		}
	}

	if op.RequestBody != nil {
		properties["body"] = op.RequestBody
		required = append(required, "body")
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	if len(properties) == 0 {
		schema["additionalProperties"] = false
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return json.RawMessage(`{"type":"object"}`)
	}
	return data
}

func extractParameters(opMap map[string]any) []Parameter {
	paramsAny, ok := opMap["parameters"].([]any)
	if !ok {
		return nil
	}
	params := make([]Parameter, 0, len(paramsAny))
	for _, pAny := range paramsAny {
		pMap, ok := pAny.(map[string]any)
		if !ok {
			continue
		}
		p := Parameter{}
		if v, ok := pMap["name"].(string); ok {
			p.Name = v
		}
		if v, ok := pMap["in"].(string); ok {
			p.In = v
		}
		if v, ok := pMap["description"].(string); ok {
			p.Description = v
		}
		if v, ok := pMap["required"].(bool); ok {
			p.Required = v
		}
		if v, ok := pMap["schema"]; ok {
			if data, err := json.Marshal(v); err == nil {
				p.Schema = data
			}
		}
		params = append(params, p)
	}
	return params
}

func extractRequestBodySchema(opMap map[string]any) json.RawMessage {
	rb, ok := opMap["requestBody"].(map[string]any)
	if !ok {
		return nil
	}
	content, ok := rb["content"].(map[string]any)
	if !ok {
		return nil
	}
	jsonContent, ok := content["application/json"].(map[string]any)
	if !ok {
		return nil
	}
	s, ok := jsonContent["schema"]
	if !ok {
		return nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return nil
	}
	return data
}
