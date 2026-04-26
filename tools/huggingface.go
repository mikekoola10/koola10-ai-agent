package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func huggingfaceTool(payload map[string]interface{}) ToolResult {
	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	switch action {
	case "search_models":
		return searchModels(payload)
	case "run_model":
		return runModel(payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func searchModels(payload map[string]interface{}) ToolResult {
	query, _ := payload["query"].(string)
	// Properly escape query for URL
	escapedQuery := url.QueryEscape(query)
	apiUrl := fmt.Sprintf("https://huggingface.co/api/models?search=%s&limit=5", escapedQuery)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("API request failed: %v", err)}
	}
	defer resp.Body.Close()

	var models []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to decode response: %v", err)}
	}

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Found %d models", len(models)),
		Data:    models,
	}
}

func runModel(payload map[string]interface{}) ToolResult {
	model, ok := payload["model"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing 'model' parameter"}
	}
	inputs := payload["inputs"]
	if inputs == nil {
		return ToolResult{Success: false, Error: "Missing 'inputs' parameter"}
	}

	token := os.Getenv("HUGGINGFACE_API_TOKEN")
	if token == "" {
		return ToolResult{Success: false, Error: "HUGGINGFACE_API_TOKEN not set"}
	}

	// Using the new router endpoint as api-inference.huggingface.co is being deprecated
	// in favor of the optimized routing layer for production workloads.
	apiUrl := fmt.Sprintf("https://router.huggingface.co/hf-inference/models/%s", model)
	body, err := json.Marshal(map[string]interface{}{"inputs": inputs})
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to marshal inputs: %v", err)}
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create request: %v", err)}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Inference failed: %v", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to read response: %v", err)}
	}

	if resp.StatusCode != http.StatusOK {
		return ToolResult{Success: false, Error: fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(respBody))}
	}

	contentType := resp.Header.Get("Content-Type")
	isImage := strings.HasPrefix(contentType, "image/")

	var outputData string
	if isImage {
		outputData = base64.StdEncoding.EncodeToString(respBody)
	} else {
		outputData = string(respBody)
	}

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Inference successful, type: %s", contentType),
		Data:    outputData,
	}
}

func init() {
	RegisterTool("huggingface", huggingfaceTool)
}
