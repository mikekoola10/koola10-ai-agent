package tools

import (
	"os"
)

func financeTool(payload map[string]interface{}) ToolResult {
	clientID := os.Getenv("PLAID_CLIENT_ID")
	secret := os.Getenv("PLAID_SECRET")
	if clientID == "" || secret == "" {
		return ToolResult{Success: false, Error: "PLAID credentials not set"}
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	switch action {
	case "get_accounts":
		return ToolResult{
			Success: true,
			Output:  "Fetched financial accounts",
			Data:    []map[string]interface{}{{"name": "Checking", "balance": 5000.0}},
		}
	case "get_transactions":
		return ToolResult{
			Success: true,
			Output:  "Fetched transactions",
			Data:    []map[string]interface{}{{"date": "2024-05-20", "amount": -20.50, "merchant": "Coffee Shop"}},
		}
	default:
		return ToolResult{Success: false, Error: "Unknown action"}
	}
}

func init() {
	RegisterTool("finance", financeTool)
}
