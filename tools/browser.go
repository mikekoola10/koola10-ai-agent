package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func browserTool(payload map[string]interface{}) ToolResult {
	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	url, _ := payload["url"].(string)
	if url == "" {
		return ToolResult{Success: false, Error: "Missing url"}
	}

	browserUrl := os.Getenv("BROWSER_AGENT_URL")
	if browserUrl == "" {
		browserUrl = "http://localhost:8081" // Default for internal use
	}

	endpoint := ""
	switch action {
	case "navigate":
		endpoint = "/browser/navigate"
	case "extract":
		endpoint = "/browser/extract"
	case "fill_form":
		endpoint = "/browser/fill-form"
	case "submit_form":
		endpoint = "/browser/submit-form"
	default:
		return ToolResult{Success: false, Error: "Unknown action"}
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(browserUrl+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Browser agent request failed: %v", err)}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Success: false, Error: "Failed to decode response"}
	}

	return ToolResult{
		Success: true,
		Data:    result,
	}
}

func init() {
	RegisterTool("browser", browserTool)
}
