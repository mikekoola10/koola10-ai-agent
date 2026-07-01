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

	email := os.Getenv("CASHAPP_EMAIL")
	password := os.Getenv("CASHAPP_PASSWORD")

	if email == "" || password == "" {
		return ToolResult{Success: false, Error: "CASHAPP_EMAIL or CASHAPP_PASSWORD not set"}
	}

	switch action {
	case "send_to_cashtag":
		return sendToCashtag(payload, browserAgentURL, email, password)
	case "get_balance":
		return getBalance(payload, browserAgentURL, email, password)
	case "get_history":
		return getHistory(payload, browserAgentURL, email, password)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func sendToCashtag(payload map[string]interface{}, browserURL, email, password string) ToolResult {
	cashtag, _ := payload["cashtag"].(string)
	amount, _ := payload["amount"].(float64)

	if cashtag == "" || amount <= 0 {
		return ToolResult{Success: false, Error: "Invalid cashtag or amount"}
	}

	task := fmt.Sprintf(`Go to https://cash.app and log in using email "%s" and password "%s".
    Once logged in, navigate to the "Pay" or "Send Money" section.
    Enter the cashtag "%s" and the amount "%.2f".
    Click the "Pay" or "Send" button and confirm the transaction.
    Provide a brief summary of the result and a confirmation message if successful.`, email, password, cashtag, amount)

	return runBrowserTask(task, browserURL)
}

func getBalance(payload map[string]interface{}, browserURL, email, password string) ToolResult {
	task := fmt.Sprintf(`Go to https://cash.app and log in using email "%s" and password "%s".
    On the home dashboard, find and extract the current balance.
    Return the balance amount.`, email, password)

	return runBrowserTask(task, browserURL)
}

func getHistory(payload map[string]interface{}, browserURL, email, password string) ToolResult {
	task := fmt.Sprintf(`Go to https://cash.app and log in using email "%s" and password "%s".
    Navigate to the activity or history section.
    Extract the most recent transactions, including amount, status, and recipient/sender.
    Return the list of transactions.`, email, password)

	return runBrowserTask(task, browserURL)
}

func runBrowserTask(task, browserURL string) ToolResult {
	reqBody, _ := json.Marshal(map[string]string{"task": task})
	resp, err := http.Post(browserURL+"/browser/task", "application/json", bytes.NewBuffer(reqBody))
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

func init() {
	RegisterTool("cashapp", cashappTool)
}
