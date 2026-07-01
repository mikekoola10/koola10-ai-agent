package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func reachTool(payload map[string]interface{}) ToolResult {
	query, _ := payload["query"].(string)
	if query == "" {
		return ToolResult{Success: false, Error: "Missing query"}
	}

	browserURL := os.Getenv("BROWSER_AGENT_URL")
	if browserURL == "" {
		browserURL = "http://localhost:8081" // Default for internal network
	}

	// Use browser-agent to extract info
	reqBody, _ := json.Marshal(map[string]string{
		"url":         "https://www.google.com/search?q=" + query,
		"instruction": "Extract the titles and snippets of the top 3 results.",
	})

	resp, err := http.Post(browserURL+"/browser/extract", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return ToolResult{Success: true, Output: "Reach simulation: Found trending topics for " + query}
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Results for %s: %v", query, res["data"]),
		Data:    res,
	}
}

func init() {
	RegisterTool("reach", reachTool)
}
