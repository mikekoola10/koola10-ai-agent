package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func cashAppTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)

	switch action {
	case "send_to_cashtag":
		cashtag, _ := payload["cashtag"].(string)
		amount, _ := payload["amount"].(float64)
		return callBrowserForCashApp(fmt.Sprintf("Send $%.2f to cashtag %s on Cash App", amount, cashtag))
	case "get_balance":
		return callBrowserForCashApp("Check my current Cash App balance")
	case "get_history":
		return callBrowserForCashApp("Get my recent Cash App transaction history")
	default:
		return ToolResult{Success: false, Error: "Invalid Cash App action"}
	}
}

func callBrowserForCashApp(task string) ToolResult {
	browserURL := os.Getenv("BROWSER_AGENT_URL")
	if browserURL == "" {
		browserURL = "http://localhost:8081"
	}

	reqBody, _ := json.Marshal(map[string]string{"task": task})
	resp, err := http.Post(browserURL+"/browser/task", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return ToolResult{Success: false, Error: "Failed to connect to browser agent: " + err.Error()}
	}
	defer resp.Body.Close()

	var res struct {
		Status      string `json:"status"`
		AgentResult string `json:"agent_result"`
		Screenshot  string `json:"screenshot"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return ToolResult{Success: false, Error: "Failed to decode browser response"}
	}

	return ToolResult{
		Success: res.Status == "success",
		Output:  res.AgentResult,
		Data:    map[string]string{"receipt_screenshot": res.Screenshot, "result": res.AgentResult},
	}
}

func init() {
	RegisterTool("cashapp", cashAppTool)
}
