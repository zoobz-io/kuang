package mcp

import (
	"encoding/json"
	"testing"
)

func TestOperationsFromSpec(t *testing.T) {
	spec := map[string]any{
		"paths": map[string]any{
			"/users": map[string]any{
				"get": map[string]any{
					"operationId": "listUsers",
					"summary":     "List all users",
				},
				"post": map[string]any{
					"operationId": "createUser",
					"summary":     "Create a user",
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"name": map[string]any{"type": "string"},
									},
								},
							},
						},
					},
				},
			},
			"/users/{id}": map[string]any{
				"get": map[string]any{
					"operationId": "getUser",
					"summary":     "Get user by ID",
					"parameters": []any{
						map[string]any{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema":   map[string]any{"type": "string"},
						},
					},
				},
			},
		},
	}

	ops := OperationsFromSpec(spec)
	if len(ops) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(ops))
	}

	byID := map[string]Operation{}
	for _, op := range ops {
		byID[op.ID] = op
	}

	if op, ok := byID["listUsers"]; !ok {
		t.Error("missing listUsers")
	} else if op.Method != "GET" {
		t.Errorf("listUsers method = %q, want GET", op.Method)
	}

	if op, ok := byID["createUser"]; !ok {
		t.Error("missing createUser")
	} else if op.RequestBody == nil {
		t.Error("createUser should have a request body")
	}

	if op, ok := byID["getUser"]; !ok {
		t.Error("missing getUser")
	} else if len(op.Parameters) != 1 {
		t.Errorf("getUser params = %d, want 1", len(op.Parameters))
	}
}

func TestOperationsFromSpecEmpty(t *testing.T) {
	ops := OperationsFromSpec(map[string]any{})
	if len(ops) != 0 {
		t.Errorf("expected 0 operations, got %d", len(ops))
	}
}

func TestToolFromOperation(t *testing.T) {
	op := Operation{
		ID:      "getUser",
		Method:  "GET",
		Path:    "/users/{id}",
		Summary: "Get user by ID",
		Parameters: []Parameter{
			{Name: "id", In: "path", Required: true, Schema: json.RawMessage(`{"type":"string"}`)},
		},
	}

	tool := ToolFromOperation(op)

	if tool.Name != "getUser" {
		t.Errorf("name = %q, want getUser", tool.Name)
	}
	if tool.Description != "Get user by ID" {
		t.Errorf("description = %q", tool.Description)
	}

	var schema map[string]any
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	props, _ := schema["properties"].(map[string]any)
	if _, ok := props["id"]; !ok {
		t.Error("schema missing 'id' property")
	}
	required, _ := schema["required"].([]any)
	if len(required) != 1 || required[0] != "id" {
		t.Errorf("required = %v, want [id]", required)
	}
}

func TestToolFromOperationNoID(t *testing.T) {
	op := Operation{
		Method: "DELETE",
		Path:   "/users/{id}",
	}
	tool := ToolFromOperation(op)
	if tool.Name != "delete_users__id_" {
		t.Errorf("name = %q, want delete_users__id_", tool.Name)
	}
}

func TestToolFromOperationWithBody(t *testing.T) {
	op := Operation{
		ID:          "createUser",
		Method:      "POST",
		Path:        "/users",
		Summary:     "Create user",
		RequestBody: json.RawMessage(`{"type":"object"}`),
	}

	tool := ToolFromOperation(op)
	var schema map[string]any
	_ = json.Unmarshal(tool.InputSchema, &schema)

	props, _ := schema["properties"].(map[string]any)
	if _, ok := props["body"]; !ok {
		t.Error("schema missing 'body' property for request body")
	}
}

func TestToolFromOperationNoParams(t *testing.T) {
	op := Operation{
		ID:      "healthCheck",
		Method:  "GET",
		Path:    "/health",
		Summary: "Health check",
	}
	tool := ToolFromOperation(op)

	var schema map[string]any
	_ = json.Unmarshal(tool.InputSchema, &schema)

	if v, ok := schema["additionalProperties"].(bool); !ok || v != false {
		t.Error("empty params should have additionalProperties: false")
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"getUser", "getUser"},
		{"get-user", "get-user"},
		{"get user", "get_user"},
		{"get/user/{id}", "get_user__id_"},
	}
	for _, tt := range tests {
		got := sanitizeName(tt.in)
		if got != tt.want {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
