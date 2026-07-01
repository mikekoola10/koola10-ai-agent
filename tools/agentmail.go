package tools

import (
	"fmt"
	"log"
)

func AgentMailTool(payload map[string]interface{}) ToolResult {
	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)

	if to == "" || subject == "" || body == "" {
		return ToolResult{Success: false, Error: "Missing required fields (to, subject, body)"}
	}

	// Mock email sending
	log.Printf("[AgentMail] Sending to %s: %s", to, subject)
	fmt.Printf("--- EMAIL START ---\nTo: %s\nSubject: %s\n\n%s\n--- EMAIL END ---\n", to, subject, body)

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"message": "Email sent successfully (mock)",
		},
	}
}

func init() {
	RegisterTool("agentmail", AgentMailTool)
}
