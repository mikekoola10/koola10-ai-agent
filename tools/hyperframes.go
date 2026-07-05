package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

func hyperframesTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	project, _ := payload["project"].(string)

	if project == "" {
		project = "default-project"
	}

	switch action {
	case "init":
		cmd := exec.Command("npx", "hyperframes", "init", "videos/"+project, "--non-interactive", "--example=blank")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("Hyperframes init failed: %v | %s", err, string(out))}
		}
		return ToolResult{Success: true, Output: string(out)}

	case "render":
		skill, _ := payload["skill"].(string)
		if skill == "" {
			skill = "general-video"
		}
		cmd := exec.Command("npx", "hyperframes", "render", "videos/"+project, "--skill="+skill, "--quality", "draft")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("Hyperframes render failed: %v | %s", err, string(out))}
		}
		return ToolResult{Success: true, Output: string(out)}

	case "status":
		cmd := exec.Command("npx", "hyperframes", "auth", "status")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("Hyperframes status failed: %v | %s", err, string(out))}
		}
		return ToolResult{Success: true, Output: string(out)}

	default:
		// Generic execution
		command, _ := payload["command"].(string)
		if command == "" {
			return ToolResult{Success: false, Error: "Missing action or command"}
		}
		args := strings.Fields(command)
		cmd := exec.Command("npx", append([]string{"hyperframes"}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("Hyperframes execution failed: %v | %s", err, string(out))}
		}
		return ToolResult{Success: true, Output: string(out)}
	}
}

func init() {
	RegisterTool("hyperframes", hyperframesTool)
}
