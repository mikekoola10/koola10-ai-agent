package tools

import (
	"fmt"
	"os"
)

func communicationTool(payload map[string]interface{}) ToolResult {
	creds := os.Getenv("COMMUNICATION_CREDENTIALS")
	if creds == "" {
		// Try MachineAuth fallback
		token, err := GetMachineAuthToken("communication-agent")
		if err == nil {
			creds = token
		} else {
			return ToolResult{Success: false, Error: "COMMUNICATION_CREDENTIALS not set and MachineAuth failed"}
		}
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	switch action {
	case "send_email":
		to, _ := payload["to"].(string)
		subject, _ := payload["subject"].(string)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Sent email to %s: %s", to, subject),
		}
	case "list_messages":
		return ToolResult{
			Success: true,
			Output:  "Listed messages",
			Data:    []string{"message1", "message2"},
		}
	case "slack_notify":
		channel, _ := payload["channel"].(string)
		message, _ := payload["message"].(string)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Sent Slack notification to %s: %s", channel, message),
		}
	default:
		return ToolResult{Success: false, Error: "Unknown action"}
	}
}

func init() {
	RegisterTool("communication", communicationTool)
}
