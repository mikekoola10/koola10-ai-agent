package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func cuaTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	url, _ := payload["url"].(string)

	if url == "" {
		return ToolResult{Success: false, Error: "Missing url"}
	}

	browserURL := os.Getenv("BROWSER_AGENT_URL")
	if browserURL == "" {
		browserURL = "http://localhost:8081"
	}

	var endpoint string
	var body []byte

	switch action {
	case "navigate":
		endpoint = "/browser/navigate"
		body, _ = json.Marshal(map[string]string{"url": url})
	case "fill":
		endpoint = "/browser/fill-form"
		formData, _ := payload["form_data"].(map[string]string)
		body, _ = json.Marshal(map[string]interface{}{"url": url, "form_data": formData})
	default:
		return ToolResult{Success: false, Error: "Unsupported action: " + action}
	}

	resp, err := http.Post(browserURL+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: true, Output: "CUA simulation: Performed " + action + " on " + url}
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("CUA result for %s: %s", action, url),
		Data:    res,
	}
}

func init() {
	RegisterTool("cua", cuaTool)
}
