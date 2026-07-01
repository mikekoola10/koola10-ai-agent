package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var emailClient = &http.Client{
	Timeout: 15 * time.Second,
}

func agentmailTool(payload map[string]interface{}) ToolResult {
	apiKey := os.Getenv("AGENTMAIL_API_KEY")
	if apiKey == "" {
		return ToolResult{Success: false, Error: "AGENTMAIL_API_KEY not set"}
	}

	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)
	if body == "" {
		body, _ = payload["text"].(string)
	}

	if to == "" || subject == "" || body == "" {
		return ToolResult{Success: false, Error: "Missing 'to', 'subject', or 'body/text'"}
	}

	inboxID := os.Getenv("AGENTMAIL_INBOX_ID")
	if inboxID == "" {
		inboxID = "mikekoola10@agentmail.to"
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"to":      to,
		"subject": subject,
		"text":    body,
	})

	url := fmt.Sprintf("https://api.agentmail.to/v0/inboxes/%s/messages/send", inboxID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := emailClient.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ToolResult{Success: false, Error: fmt.Sprintf("API returned status %d", resp.StatusCode)}
	}

	return ToolResult{Success: true, Output: "Email sent successfully"}
}

func init() {
	RegisterTool("agentmail", agentmailTool)
	RegisterTool("send_email", agentmailTool)
}
