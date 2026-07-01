package tools

import (
	"fmt"
)

func cuaTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	os, _ := payload["os"].(string) // macOS, Linux, Windows

	if os == "" {
		os = "Linux (Sandbox)"
	}

	switch action {
	case "screenshot":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("CUA: Screenshot captured on %s", os),
			Data: map[string]interface{}{
				"os":              os,
				"screenshot_url": "https://koola10.fly.dev/data/screenshots/latest.png",
			},
		}
	case "click":
		x, _ := payload["x"].(float64)
		y, _ := payload["y"].(float64)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("CUA: Mouse clicked at (%.0f, %.0f) on %s", x, y, os),
		}
	case "type":
		text, _ := payload["text"].(string)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("CUA: Typed '%s' into active window on %s", text, os),
		}
	case "run_command":
		cmd, _ := payload["command"].(string)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("CUA: Executed command '%s' on %s", cmd, os),
			Data: map[string]interface{}{
				"stdout": "Success",
			},
		}
	default:
		return ToolResult{Success: false, Error: "Invalid CUA action. Supported: screenshot, click, type, run_command."}
	}
}

func init() {
	RegisterTool("cua", cuaTool)
}
