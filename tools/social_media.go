package tools

import (
	"fmt"
	"os"
)

func socialMediaTool(payload map[string]interface{}) ToolResult {
	tokens := os.Getenv("SOCIAL_MEDIA_TOKENS")
	if tokens == "" {
		return ToolResult{Success: false, Error: "SOCIAL_MEDIA_TOKENS not set"}
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	platform, _ := payload["platform"].(string) // linkedin, twitter, facebook, instagram

	switch action {
	case "post_content":
		content, _ := payload["content"].(string)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Posted content to %s: %s", platform, content),
			Data:    map[string]string{"post_id": "post_789"},
		}
	case "get_metrics":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Fetched metrics for %s", platform),
			Data:    map[string]int{"likes": 120, "shares": 15},
		}
	default:
		return ToolResult{Success: false, Error: "Unknown action"}
	}
}

func init() {
	RegisterTool("social_media", socialMediaTool)
}
