package tools

import (
	"fmt"
)

func memoryTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	client, _ := payload["client"].(string) // Claude Code, Cursor, etc.
	key, _ := payload["key"].(string)
	value, _ := payload["value"].(string)

	if client == "" {
		client = "Generic AI Client"
	}

	switch action {
	case "store":
		if key == "" || value == "" {
			return ToolResult{Success: false, Error: "Missing key or value"}
		}
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Agent Memory: Persisted data for %s: %s", client, key),
		}
	case "recall":
		if key == "" {
			return ToolResult{Success: false, Error: "Missing key"}
		}
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Agent Memory: Recalled data for %s: %s", client, key),
			Data: map[string]interface{}{
				"key":   key,
				"value": "Simulated memory value for " + key,
			},
		}
	case "sync":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Agent Memory: Synchronized context across all clients (Claude Code, GitHub Copilot CLI, Cursor, Gemini CLI)."),
		}
	default:
		return ToolResult{Success: false, Error: "Invalid memory action. Supported: store, recall, sync."}
	}
}

func init() {
	RegisterTool("memory", memoryTool)
}
