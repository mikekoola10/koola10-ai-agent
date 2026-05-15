package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// openComputerUseTool handles desktop automation via an Open Computer Use instance.
// It supports actions: click, type, scroll, screenshot, run_command, open_app.
func openComputerUseTool(payload map[string]interface{}) ToolResult {
	baseUrl := os.Getenv("OPEN_COMPUTER_USE_URL")
	if baseUrl == "" {
		baseUrl = "http://koola10-desktop.fly.dev"
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	// Verify supported actions
	supported := false
	for _, a := range []string{"click", "type", "scroll", "screenshot", "run_command", "open_app"} {
		if action == a {
			supported = true
			break
		}
	}
	if !supported {
		return ToolResult{Success: false, Error: fmt.Sprintf("Unsupported action: %s", action)}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to marshal payload: %v", err)}
	}

	// Forward the request to the Open Computer Use service
	resp, err := http.Post(baseUrl+"/execute", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to connect to Open Computer Use service at %s: %v", baseUrl, err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to read response from Open Computer Use service: %v", err)}
	}

	if resp.StatusCode != http.StatusOK {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Open Computer Use service returned error (status %d): %s", resp.StatusCode, string(respBody)),
		}
	}

	var result ToolResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Fallback if the service doesn't return a standard ToolResult JSON
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Desktop action '%s' executed successfully", action),
			Data:    string(respBody),
		}
	}

	return result
}

func init() {
	RegisterTool("computeruse", openComputerUseTool)
}
