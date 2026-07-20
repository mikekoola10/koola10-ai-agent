package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func strikeTool(payload map[string]interface{}) ToolResult {
	apiKey := os.Getenv("STRIKE_API_KEY")
	if apiKey == "" {
		return ToolResult{Success: false, Error: "STRIKE_API_KEY not set"}
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	switch action {
	case "get_balance":
		return strikeGet(apiKey, "https://api.strike.me/v1/balances", "Balance retrieved successfully")
	case "get_payment_history":
		return strikeGet(apiKey, "https://api.strike.me/v1/payments", "Payment history retrieved successfully")
	case "get_exchange_rates":
		return strikeGet(apiKey, "https://api.strike.me/v1/rates/ticker", "Exchange rates retrieved successfully")
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func strikeGet(apiKey, url, successMsg string) ToolResult {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ToolResult{Success: false, Error: fmt.Sprintf("Strike API error: %s", resp.Status)}
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to decode response: %v", err)}
	}

	return ToolResult{
		Success: true,
		Output:  successMsg,
		Data:    data,
	}
}

func init() {
	RegisterTool("strike", strikeTool)
}
