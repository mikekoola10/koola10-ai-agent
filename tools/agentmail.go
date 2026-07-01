package tools

import (
	"fmt"
	"log"
)

func agentmailTool(payload map[string]interface{}) ToolResult {
	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	switch action {
	case "send_email":
		return sendEmail(payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func sendEmail(payload map[string]interface{}) ToolResult {
	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)

	if to == "" || body == "" {
		return ToolResult{Success: false, Error: "Missing 'to' or 'body' parameter"}
	}

	// In a real implementation, this would call an external API.
	// For this simulation, we'll log the email and return success.
	log.Printf("[AgentMail] Sending email to: %s, Subject: %s, Body: %s", to, subject, body)

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Email sent to %s", to),
	}
}

func init() {
	RegisterTool("agentmail", agentmailTool)
}
