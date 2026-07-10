package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func init() {
	RegisterTool("anthropic", AnthropicTool)
}

func AnthropicTool(payload map[string]interface{}) ToolResult {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return ToolResult{Success: false, Error: "ANTHROPIC_API_KEY not set"}
	}

	prompt, _ := payload["prompt"].(string)
	system, _ := payload["system"].(string)
	model := "claude-3-5-sonnet-20240620"

	reqBody := map[string]interface{}{
		"model":      model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	if system != "" {
		reqBody["system"] = system
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return ToolResult{Success: false, Error: fmt.Sprintf("API error: %s", string(respBody))}
	}

	var res struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &res); err != nil {
		return ToolResult{Success: false, Error: "Failed to parse response"}
	}

	if len(res.Content) > 0 {
		return ToolResult{Success: true, Output: res.Content[0].Text}
	}

	return ToolResult{Success: false, Error: "No content in response"}
}
