package tools

import (
	"fmt"
)

func agentMailTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)

	switch action {
	case "send":
		to, _ := payload["to"].(string)
		return ToolResult{Success: true, Output: fmt.Sprintf("Email sent to %s via AgentMail", to)}
	case "list":
		return ToolResult{Success: true, Data: map[string]interface{}{"emails": []string{"Welcome to Koola10", "System Alert"}}}
	default:
		return ToolResult{Success: false, Error: "Invalid action"}
	}
}

func init() {
	RegisterTool("agentmail", agentMailTool)
}
