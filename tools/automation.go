package tools

import (
	"fmt"
)

func automationTool(payload map[string]interface{}) ToolResult {
	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	platform, _ := payload["platform"].(string) // zapier, make, n8n

	switch action {
	case "trigger_webhook":
		url, _ := payload["webhook_url"].(string)
		data, _ := payload["data"].(map[string]interface{})
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Triggered %s webhook: %s", platform, url),
			Data:    data,
		}
	case "list_automations":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Listed automations for %s", platform),
			Data:    []string{"automation_A", "automation_B"},
		}
	default:
		return ToolResult{Success: false, Error: "Unknown action"}
	}
}

func init() {
	RegisterTool("automation", automationTool)
}
