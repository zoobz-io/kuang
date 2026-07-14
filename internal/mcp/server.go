package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// Backend defines the operations the server needs from a kuang API client.
type Backend interface {
	FetchSpec(ctx context.Context) (map[string]any, error)
	Call(ctx context.Context, op Operation, arguments json.RawMessage) (int, []byte, error)
}

// Server is an MCP server that bridges OpenAPI operations over stdio.
type Server struct {
	backend Backend
	tools   []Tool
	opMap   map[string]Operation // tool name → operation
	ops     []Operation
}

// NewServer creates a server that will fetch tools from the given backend.
func NewServer(backend Backend) *Server {
	return &Server{
		backend: backend,
		opMap:   make(map[string]Operation),
	}
}

// ToolCount returns the number of loaded tools.
func (s *Server) ToolCount() int { return len(s.tools) }

// LoadTools fetches the OpenAPI spec and builds the tool catalog.
func (s *Server) LoadTools() error {
	spec, err := s.backend.FetchSpec(context.Background())
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}

	s.ops = OperationsFromSpec(spec)
	s.tools = make([]Tool, len(s.ops))
	s.opMap = make(map[string]Operation, len(s.ops))

	for i, op := range s.ops {
		tool := ToolFromOperation(op)
		s.tools[i] = tool
		s.opMap[tool.Name] = op
	}

	return nil
}

// Run reads JSON-RPC messages from r and writes responses to w.
func (s *Server) Run(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("mcp: invalid json-rpc: %v", err)
			continue
		}

		resp := s.handle(req)
		if resp == nil {
			continue // notification, no response
		}

		data, err := json.Marshal(resp)
		if err != nil {
			log.Printf("mcp: marshal response: %v", err)
			continue
		}
		data = append(data, '\n')
		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}

	return scanner.Err()
}

func (s *Server) handle(req Request) *Response {
	switch req.Method {
	case "initialize":
		return s.respond(req.ID, InitializeResult{
			ProtocolVersion: protocolVersion,
			Capabilities: Capabilities{
				Tools: &ToolsCap{},
			},
			ServerInfo: ServerInfo{
				Name:    "kuang-mcp",
				Version: "0.1.0",
			},
		})

	case "notifications/initialized":
		return nil // notification, no response

	case "tools/list":
		return s.respond(req.ID, ToolsListResult{
			Tools: s.tools,
		})

	case "tools/call":
		return s.handleToolCall(req)

	case "ping":
		return s.respond(req.ID, map[string]string{})

	default:
		return s.respondError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *Server) handleToolCall(req Request) *Response {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.respondError(req.ID, -32602, "invalid params")
	}

	op, ok := s.opMap[params.Name]
	if !ok {
		return s.respondError(req.ID, -32602, fmt.Sprintf("unknown tool: %s", params.Name))
	}

	status, body, err := s.backend.Call(context.Background(), op, params.Arguments)
	if err != nil {
		return s.respond(req.ID, ToolCallResult{
			Content: []Content{{Type: "text", Text: fmt.Sprintf("request failed: %v", err)}},
			IsError: true,
		})
	}

	isError := status >= 400
	text := string(body)
	if text == "" {
		text = fmt.Sprintf("HTTP %d", status)
	}

	return s.respond(req.ID, ToolCallResult{
		Content: []Content{{Type: "text", Text: text}},
		IsError: isError,
	})
}

func (s *Server) respond(id json.RawMessage, result any) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func (s *Server) respondError(id json.RawMessage, code int, message string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
}
