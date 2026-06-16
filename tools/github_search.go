package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type GitHubRepo struct {
	Name        string `json:"full_name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	URL         string `json:"html_url"`
}

func githubSearchTool(payload map[string]interface{}) ToolResult {
	query, _ := payload["query"].(string)
	if query == "" {
		return ToolResult{Success: false, Error: "Missing query"}
	}

	searchURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&order=desc", url.QueryEscape(query))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	req.Header.Set("User-Agent", "Koola10-Agent")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	var result struct {
		Items []GitHubRepo `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Success: false, Error: "Failed to decode response"}
	}

	limit := 5
	if len(result.Items) < limit {
		limit = len(result.Items)
	}

	return ToolResult{
		Success: true,
		Data:    map[string]interface{}{"repos": result.Items[:limit]},
	}
}

func init() {
	RegisterTool("github_search", githubSearchTool)
}
