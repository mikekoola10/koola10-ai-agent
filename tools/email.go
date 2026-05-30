package tools

import (
	"fmt"
)

func emailTool(payload map[string]interface{}) ToolResult {
	recipient, _ := payload["recipient"].(string)
	company, _ := payload["company"].(string)
	body, _ := payload["body"].(string)

	if recipient == "" || body == "" {
		return ToolResult{Success: false, Error: "Missing recipient or body"}
	}

	// Simulation of sending email
	fmt.Printf("--- SIMULATED EMAIL SENT ---\nTo: %s\nCompany: %s\nBody: %s\n---------------------------\n", recipient, company, body)

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Email sent to %s", recipient),
		Data:    map[string]interface{}{"recipient": recipient, "company": company, "status": "sent"},
	}
}

func init() {
	RegisterTool("email", emailTool)
}
