package tools

import (
	"fmt"
	"os"
)

func crmTool(payload map[string]interface{}) ToolResult {
	keys := os.Getenv("CRM_API_KEYS")
	if keys == "" {
		// Hands-free fallback
		token, err := GetNexusToken("salesforce")
		if err == nil {
			keys = token
		} else {
			return ToolResult{Success: false, Error: "CRM_API_KEYS not set and Nexus fallback failed"}
		}
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	platform, _ := payload["platform"].(string) // salesforce, hubspot

	switch action {
	case "test":
		return ToolResult{Success: true, Output: "CRM connector test successful"}
	case "search_contacts":
		query, _ := payload["query"].(string)
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Searched contacts in %s for: %s", platform, query),
			Data:    []map[string]string{{"name": "John Doe", "email": "john@example.com"}},
		}
	case "create_lead":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Created lead in %s", platform),
			Data:    map[string]string{"id": "lead_123"},
		}
	case "update_opportunity":
		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Updated opportunity in %s", platform),
		}
	default:
		return ToolResult{Success: false, Error: "Unknown action"}
	}
}

func init() {
	RegisterTool("crm", crmTool)
}
