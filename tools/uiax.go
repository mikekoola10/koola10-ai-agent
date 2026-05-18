package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// MCPRequest represents a JSON-RPC 2.0 request for the Model Context Protocol
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

func uiaxTool(payload map[string]interface{}) ToolResult {
	serverURL := os.Getenv("UIAX_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8000" // Default for local dev, but should be set for Fly.io
	}
	apiKey := os.Getenv("UIAX_API_KEY")

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	switch action {
	case "screenshot":
		return callMCPTool(serverURL, apiKey, "screenshot", nil)
	case "read_window":
		windowName, _ := payload["window_name"].(string)
		return callMCPTool(serverURL, apiKey, "read_window", map[string]interface{}{"window_title": windowName})
	case "click_element":
		elementName, _ := payload["element_name"].(string)
		return callMCPTool(serverURL, apiKey, "click_element", map[string]interface{}{"name": elementName})
	case "type_text":
		text, _ := payload["text"].(string)
		return callMCPTool(serverURL, apiKey, "type_text", map[string]interface{}{"text": text})
	case "clipboard_copy":
		windowName, _ := payload["window_name"].(string)
		return callMCPTool(serverURL, apiKey, "clipboard_copy", map[string]interface{}{"window_title": windowName})
	case "clipboard_paste":
		return callMCPTool(serverURL, apiKey, "clipboard_paste", nil)
	case "watch_conversation":
		return watchConversation(serverURL, apiKey, payload)
	case "switch_to_jules":
		return switchToJules(serverURL, apiKey, payload)
	case "open_terminal":
		return callMCPTool(serverURL, apiKey, "open_terminal", map[string]interface{}{"shell": "powershell"})
	case "run_command":
		command, _ := payload["command"].(string)
		return callMCPTool(serverURL, apiKey, "run_command", map[string]interface{}{"command": command})
	case "deploy_via_terminal":
		return callMCPTool(serverURL, apiKey, "run_command", map[string]interface{}{"command": "flyctl deploy --app koola10 --dockerfile ./Dockerfile"})
	case "vercel_deploy_via_terminal":
		token := os.Getenv("VERCEL_TOKEN")
		command := fmt.Sprintf("cd ceo-dashboard; $env:VERCEL_TOKEN='%s'; npx vercel --prod --token $env:VERCEL_TOKEN --yes", token)
		return callMCPTool(serverURL, apiKey, "run_command", map[string]interface{}{"command": command})
	case "focus_window":
		windowName, _ := payload["window_name"].(string)
		if windowName == "" { windowName, _ = payload["window_title"].(string) }
		return callMCPTool(serverURL, apiKey, "focus_window", map[string]interface{}{"window_title": windowName})
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func callMCPTool(serverURL, apiKey, toolName string, args map[string]interface{}) ToolResult {
	if args == nil {
		args = make(map[string]interface{})
	}

	mcpReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1, // Simplified ID for now
		Method:  "tools/call",
		Params: CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}

	body, err := json.Marshal(mcpReq)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to marshal MCP request: %v", err)}
	}

	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create request: %v", err)}
	}

	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("request to UIA-X failed: %v", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to read response: %v", err)}
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(respBody, &mcpResp); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to decode MCP response: %v. Body: %s", err, string(respBody))}
	}

	if mcpResp.Error != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("MCP Error (%d): %s", mcpResp.Error.Code, mcpResp.Error.Message)}
	}

	// The result of tools/call is expected to be a CallToolResult
	return ToolResult{
		Success: true,
		Output:  string(mcpResp.Result),
		Data:    mcpResp.Result,
	}
}

func watchConversation(serverURL, apiKey string, payload map[string]interface{}) ToolResult {
	windowName, _ := payload["window_name"].(string)
	if windowName == "" {
		return ToolResult{Success: false, Error: "missing window_name"}
	}

	// 1. Read window content or use clipboard_copy
	readRes := callMCPTool(serverURL, apiKey, "read_window", map[string]interface{}{"window_title": windowName})
	if !readRes.Success {
		return readRes
	}

	copyRes := callMCPTool(serverURL, apiKey, "clipboard_copy", map[string]interface{}{"window_title": windowName})
	if !copyRes.Success {
		return copyRes
	}

	return ToolResult{Success: true, Output: "Conversation watched and text copied to clipboard", Data: readRes.Data}
}

func switchToJules(serverURL, apiKey string, payload map[string]interface{}) ToolResult {
	targetWindow := payload["target_window"].(string)
	if targetWindow == "" {
		targetWindow = "Jules"
	}

	// 1. Focus Jules window
	focusRes := callMCPTool(serverURL, apiKey, "focus_window", map[string]interface{}{"window_title": targetWindow})
	if !focusRes.Success {
		// Try fallback if focus_window doesn't exist
		focusRes = callMCPTool(serverURL, apiKey, "click_element", map[string]interface{}{"name": targetWindow})
	}

	// 2. Paste
	pasteRes := callMCPTool(serverURL, apiKey, "clipboard_paste", nil)
	if !pasteRes.Success {
		return pasteRes
	}

	// 3. Submit (e.g. press Enter)
	submitRes := callMCPTool(serverURL, apiKey, "type_text", map[string]interface{}{"text": "\n"})
	if !submitRes.Success {
		return submitRes
	}

	return ToolResult{Success: true, Output: "Switched to Jules and pasted content"}
}

func init() {
	RegisterTool("uiax", uiaxTool)
}
