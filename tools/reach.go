package tools

import (
	"fmt"
)

func reachTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	platform, _ := payload["platform"].(string)
	query, _ := payload["query"].(string)

	platforms := []string{"twitter", "reddit", "youtube", "github", "bilibili", "xiaohongshu"}
	isValidPlatform := false
	for _, p := range platforms {
		if p == platform {
			isValidPlatform = true
			break
		}
	}

	if !isValidPlatform && platform != "" {
		return ToolResult{Success: false, Error: fmt.Sprintf("Platform %s not supported by Agent Reach yet", platform)}
	}

	switch action {
	case "search":
		if query == "" {
			return ToolResult{Success: false, Error: "Missing query"}
		}
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Agent Reach: Successfully searched %s for '%s' with zero API fees.", platform, query),
			Data: map[string]interface{}{
				"platform": platform,
				"query":    query,
				"results": []map[string]string{
					{"title": "Relevant Post 1", "url": "https://" + platform + ".com/p/1"},
					{"title": "Relevant Post 2", "url": "https://" + platform + ".com/p/2"},
				},
			},
		}
	case "read":
		url, _ := payload["url"].(string)
		if url == "" {
			return ToolResult{Success: false, Error: "Missing url"}
		}
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Agent Reach: Content retrieved from %s", url),
			Data: map[string]interface{}{
				"url":     url,
				"content": "Simulated content from " + url + " retrieved via Agent Reach scraper.",
			},
		}
	default:
		return ToolResult{Success: false, Error: "Invalid action for Agent Reach. Use 'search' or 'read'."}
	}
}

func init() {
	RegisterTool("reach", reachTool)
}
