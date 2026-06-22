package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func mcpRealEstateTool(payload map[string]interface{}) ToolResult {
	mcpURL := os.Getenv("MCP_HUB_URL")
	if mcpURL == "" {
		mcpURL = "http://localhost:8090" // Default MCP Hub
	}

	action, _ := payload["action"].(string)
	params := payload["params"]

	url := fmt.Sprintf("%s/realestate/%s", mcpURL, action)
	body, _ := json.Marshal(params)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	return ToolResult{
		Success: true,
		Data:    res,
		Output:  fmt.Sprintf("MCP Real Estate %s executed", action),
	}
}

func init() {
	RegisterTool("mcp_realestate", mcpRealEstateTool)
}
