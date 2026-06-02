package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func cashappTool(payload map[string]interface{}) ToolResult {
	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	browserAgentURL := os.Getenv("BROWSER_AGENT_URL")
	if browserAgentURL == "" {
		browserAgentURL = "http://koola10-browser.fly.dev"
	}

	switch action {
	case "send_to_cashtag":
		return sendToCashtag(payload, browserAgentURL)
	case "get_balance":
		return getBalance(payload, browserAgentURL)
	case "get_history":
		return getHistory(payload, browserAgentURL)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func sendToCashtag(payload map[string]interface{}, browserURL string) ToolResult {
	cashtag, _ := payload["cashtag"].(string)
	amount, _ := payload["amount"].(float64)

	if cashtag == "" || amount <= 0 {
		return ToolResult{Success: false, Error: "Invalid cashtag or amount"}
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"cashtag": cashtag,
		"amount":  amount,
	})
	resp, err := http.Post(browserURL+"/browser/cashapp/send", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to contact browser agent: %v", err)}
	}
	defer resp.Body.Close()

	var result struct {
		Status     string `json:"status"`
		Result     string `json:"result"`
		Screenshot string `json:"screenshot"`
		Error      string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to parse browser agent response: %v", err)}
	}

	if result.Status != "success" {
		return ToolResult{Success: false, Error: result.Error, Output: result.Result}
	}

	return ToolResult{
		Success: true,
		Output:  result.Result,
		Data:    map[string]string{"screenshot": result.Screenshot},
	}
}

func getBalance(payload map[string]interface{}, browserURL string) ToolResult {
	resp, err := http.Get(browserURL + "/browser/cashapp/balance")
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to contact browser agent: %v", err)}
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
		Result string `json:"result"`
		Error  string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to parse browser agent response: %v", err)}
	}

	if result.Status != "success" {
		return ToolResult{Success: false, Error: result.Error, Output: result.Result}
	}

	return ToolResult{
		Success: true,
		Output:  result.Result,
	}
}

func getHistory(payload map[string]interface{}, browserURL string) ToolResult {
	resp, err := http.Get(browserURL + "/browser/cashapp/history")
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to contact browser agent: %v", err)}
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
		Result string `json:"result"`
		Error  string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to parse browser agent response: %v", err)}
	}

	if result.Status != "success" {
		return ToolResult{Success: false, Error: result.Error, Output: result.Result}
	}

	return ToolResult{
		Success: true,
		Output:  result.Result,
	}
}

func init() {
	RegisterTool("cashapp", cashappTool)
}
