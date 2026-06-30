package tools
import (
	"fmt"
	"log"
)
func agentMailTool(payload map[string]interface{}) ToolResult {
	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)
	if to == "" || subject == "" || body == "" {
		return ToolResult{Success: false, Error: "Missing required fields (to, subject, body)"}
	}
	log.Printf("[AgentMail] Sending to: %s\nSubject: %s\nBody: %s", to, subject, body)
	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Email queued for %s", to),
	}
}
func init() {
	RegisterTool("agentmail", agentMailTool)
}
