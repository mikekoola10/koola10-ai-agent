package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func init() {
	RegisterTool("agentmail", SendAgentMail)
}

func SendAgentMail(payload map[string]interface{}) ToolResult {
	apiKey := os.Getenv("AGENTMAIL_API_KEY")
	if apiKey == "" {
		return ToolResult{Success: false, Error: "AGENTMAIL_API_KEY not set"}
	}

	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)

	if to == "" || subject == "" || body == "" {
		return ToolResult{Success: false, Error: "missing parameters (to, subject, body)"}
	}

	mailReq := map[string]string{
		"to":      to,
		"subject": subject,
		"body":    body,
	}
	jsonData, _ := json.Marshal(mailReq)

	req, err := http.NewRequest("POST", "https://api.agentmail.to/v1/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: "failed to send email: " + err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return ToolResult{Success: false, Error: fmt.Sprintf("AgentMail API returned %d", resp.StatusCode)}
	}

	return ToolResult{
		Success: true,
		Data:    map[string]string{"status": "sent", "to": to},
	}
}
