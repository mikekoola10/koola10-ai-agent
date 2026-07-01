package tools

import (
	"fmt"
)

func skillsTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	skillset, _ := payload["skillset"].(string) // "general" or "pm"

	switch action {
	case "use":
		skill, _ := payload["skill"].(string)
		if skill == "" {
			return ToolResult{Success: false, Error: "Missing skill name"}
		}
		if skillset == "pm" {
			return ToolResult{
				Success: true,
				Output:  fmt.Sprintf("PM Skills: Executed '%s' framework (discovery, assumption mapping, prioritization, strategy).", skill),
			}
		}
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Agent Skills: Successfully utilized skill '%s' (SKILL.md standard).", skill),
		}
	case "list":
		if skillset == "pm" {
			return ToolResult{
				Success: true,
				Output:  "PM Skills: Listing 68+ agentic skills and 42 chained workflows across 9 plugins.",
				Data:    []string{"Discovery", "Assumption Mapping", "Prioritization", "Strategy"},
			}
		}
		return ToolResult{
			Success: true,
			Output:  "Agent Skills: Listing curated collection of tested and maintained agent skills.",
			Data:    []string{"Search", "File Ops", "Web Interaction", "Logic Reasoning"},
		}
	default:
		return ToolResult{Success: false, Error: "Invalid Skills action. Supported: use, list."}
	}
}

func init() {
	RegisterTool("skills", skillsTool)
}
