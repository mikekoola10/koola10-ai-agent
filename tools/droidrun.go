package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// droidrunTool handles mobile device automation via a DroidRun instance.
// It supports actions: tap, swipe, type, launch_app, screenshot, get_screen_content.
func droidrunTool(payload map[string]interface{}) ToolResult {
	baseUrl := os.Getenv("DROIDRUN_URL")
	if baseUrl == "" {
		baseUrl = "http://koola10-droidrun.fly.dev"
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	// Verify supported actions (optional but good for validation)
	supported := false
	for _, a := range []string{"tap", "swipe", "type", "launch_app", "screenshot", "get_screen_content"} {
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

	// Forward the request to the DroidRun service
	resp, err := http.Post(baseUrl+"/run", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to connect to DroidRun service at %s: %v", baseUrl, err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to read response from DroidRun service: %v", err)}
	}

	if resp.StatusCode != http.StatusOK {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("DroidRun service returned error (status %d): %s", resp.StatusCode, string(respBody)),
		}
	}

	var result ToolResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Fallback if the service doesn't return a standard ToolResult JSON
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Action '%s' executed successfully", action),
			Data:    string(respBody),
		}
	}

	return result
}

func init() {
	RegisterTool("droidrun", droidrunTool)
}
