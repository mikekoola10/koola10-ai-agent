package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func init() {
	RegisterTool("github_search", GitHubSearch)
}

func GitHubSearch(payload map[string]interface{}) ToolResult {
	query, ok := payload["query"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "missing query parameter"}
	}

	searchURL := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&order=desc", url.QueryEscape(query))
	resp, err := http.Get(searchURL)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			HTMLURL     string `json:"html_url"`
			Description string `json:"description"`
			Stars       int    `json:"stargazers_count"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Success: false, Error: "failed to parse github response"}
	}

	return ToolResult{
		Success: true,
		Data:    result.Items,
	}
}
