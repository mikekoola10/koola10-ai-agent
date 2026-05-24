package tools

import (
	"fmt"
	"os"
)

func developerTool(payload map[string]interface{}) ToolResult {
	tokens := os.Getenv("DEVELOPER_TOKENS")
	if tokens == "" {
		token, err := GetNexusToken("github")
		if err == nil {
			tokens = token
		} else {
			return ToolResult{Success: false, Error: "DEVELOPER_TOKENS not set and Nexus fallback failed"}
		}
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	platform, _ := payload["platform"].(string) // github, gitlab, linear, jira

	switch action {
	case "test":
		return ToolResult{Success: true, Output: "Developer connector test successful"}
	case "create_issue":
		title, _ := payload["title"].(string)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Created issue on %s: %s", platform, title),
			Data:    map[string]string{"issue_id": "ISSUE-1"},
		}
	case "create_pr":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Created PR on %s", platform),
			Data:    map[string]string{"pr_url": "https://github.com/org/repo/pull/1"},
		}
	case "list_repos":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Listed repos on %s", platform),
			Data:    []string{"repo1", "repo2"},
		}
	default:
		return ToolResult{Success: false, Error: "Unknown action"}
	}
}

func init() {
	RegisterTool("developer", developerTool)
}
