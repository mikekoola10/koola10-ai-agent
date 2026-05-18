package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func init() {
	RegisterTool("deepseek", deepseekTool)
}

func deepseekTool(payload map[string]interface{}) ToolResult {
	prompt, _ := payload["prompt"].(string)
	if prompt == "" {
		return ToolResult{Success: false, Error: "Missing prompt"}
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		// Mock response if no API key is present for development
		return ToolResult{
			Success: true,
			Output:  "Mock DeepSeek response (No API Key)",
			Data: map[string]interface{}{
				"content": fmt.Sprintf("Drafted content for: %s", prompt),
			},
		}
	}

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	dsBody, _ := json.Marshal(dsReq)
	hReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(dsBody))
	hReq.Header.Set("Authorization", "Bearer "+apiKey)
	hReq.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(hReq)
	if err != nil {
		return ToolResult{Success: false, Error: "API call failed: " + err.Error()}
	}
	defer resp.Body.Close()

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
		Usage struct {
			TotalTokens int
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		return ToolResult{Success: false, Error: "Parse failed: " + err.Error()}
	}

	if len(dsRes.Choices) == 0 {
		return ToolResult{Success: false, Error: "No choices returned from DeepSeek"}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"content":      dsRes.Choices[0].Message.Content,
			"tokens_used": dsRes.Usage.TotalTokens,
		},
	}
}
