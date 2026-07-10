package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MCPClient struct {
	Name string
	URL  string
}

func NewMCPClient(name string, args ...string) (*MCPClient, error) {
	// Simple implementation for HTTP transport based on user prompt
	url := ""
	for i, arg := range args {
		if arg == "--transport" && i+1 < len(args) && args[i+1] == "http" {
			if i+2 < len(args) {
				url = args[i+2]
			}
			break
		}
	}
	if url == "" {
		// Fallback: take the last argument if it looks like a URL
		if len(args) > 0 {
			url = args[len(args)-1]
		}
	}

	if url == "" {
		return nil, fmt.Errorf("no URL provided for MCP client")
	}

	return &MCPClient{
		Name: name,
		URL:  url,
	}, nil
}

func (c *MCPClient) Call(method string, params map[string]interface{}) (interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      method,
			"arguments": params,
		},
		"id": 1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	content := string(respBody)
	// If it's SSE, we need to extract the data part
	if strings.Contains(content, "event: message") || strings.HasPrefix(content, "data:") {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "data:") {
				dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				var jsonResp struct {
					Result interface{} `json:"result"`
					Error  interface{} `json:"error"`
				}
				if err := json.Unmarshal([]byte(dataStr), &jsonResp); err == nil {
					if jsonResp.Error != nil {
						return nil, fmt.Errorf("MCP error: %v", jsonResp.Error)
					}
					return jsonResp.Result, nil
				}
			}
		}
	}

	var jsonResp struct {
		Result interface{} `json:"result"`
		Error  interface{} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &jsonResp); err != nil {
		return nil, fmt.Errorf("Failed to decode MCP response: %v | Raw: %s", err, content)
	}

	if jsonResp.Error != nil {
		return nil, fmt.Errorf("MCP error: %v", jsonResp.Error)
	}

	return jsonResp.Result, nil
}
