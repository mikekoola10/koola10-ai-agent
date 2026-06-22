package tools

import (
	"fmt"
	"log"
)

func agentMailTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)

	if action == "" {
		action = "send"
	}

	switch action {
	case "send":
		if to == "" {
			return ToolResult{Success: false, Error: "missing 'to' parameter"}
		}
		// In a real implementation, this would call an email API.
		// For now, we log it and return success.
		log.Printf("[AgentMail] Sending email to %s: %s\nBody: %s", to, subject, body)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Email sent to %s with subject: %s", to, subject),
		}
	default:
		return ToolResult{Success: false, Error: "invalid action"}
	}
}

func init() {
	RegisterTool("agentmail", agentMailTool)
}
