package tools

import (
	"fmt"
)

func init() {
	RegisterTool("email", EmailTool)
}

func EmailTool(payload map[string]interface{}) ToolResult {
	to, ok := payload["to"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing 'to' field"}
	}
	subject, ok := payload["subject"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing 'subject' field"}
	}
	body, ok := payload["body"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing 'body' field"}
	}

	// Simulation: Log the email to the console and return success
	fmt.Printf("[Email Tool] Sending to: %s\nSubject: %s\nBody: %s\n", to, subject, body)

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"message": "Email sent successfully (simulated)",
			"to":      to,
			"subject": subject,
		},
	}
}
