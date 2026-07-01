package tools

import (
	"fmt"
)

func messagingTool(payload map[string]interface{}) ToolResult {
	channel, _ := payload["channel"].(string) // "slack", "sms"
	message, _ := payload["message"].(string)

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Message sent to %s: %s", channel, message),
	}
}

func init() {
	RegisterTool("messaging", messagingTool)
}
