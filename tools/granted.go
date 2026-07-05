package tools

import (
	"fmt"
	"log"
)

var grantedClient *MCPClient

func init() {
	var err error
	grantedClient, err = NewMCPClient(
		"granted",
		"--transport", "http",
		"https://grantedai.com/api/mcp/mcp",
	)
	if err != nil {
		log.Printf("Warning: failed to initialize Granted MCP client: %v", err)
		return
	}

	toolNames := []string{
		"search_grants",
		"get_grant",
		"search_funders",
		"get_funder",
		"get_past_winners",
	}

	for _, name := range toolNames {
		toolName := name // capture for closure
		RegisterTool(toolName, func(args map[string]interface{}) ToolResult {
			if grantedClient == nil {
				return ToolResult{Success: false, Error: "Granted client not initialized"}
			}
			result, err := grantedClient.Call(toolName, args)
			if err != nil {
				return ToolResult{Success: false, Error: fmt.Sprintf("Granted MCP call failed: %v", err)}
			}
			return ToolResult{
				Success: true,
				Output:  fmt.Sprintf("Result from %s", toolName),
				Data:    result,
			}
		})
	}
}
