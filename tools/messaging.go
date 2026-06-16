package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func init() {
	RegisterTool("messaging", UnifiedMessaging)
}

func SendSlackMessage(webhookURL, message string) error {
	payload := map[string]string{"text": message}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil { return err }
	defer resp.Body.Close()
	return nil
}

func UnifiedMessaging(payload map[string]interface{}) ToolResult {
	channel, _ := payload["channel"].(string) // "slack", "sms"
	message, _ := payload["message"].(string)

	if channel == "slack" {
		url := os.Getenv("SLACK_WEBHOOK_URL")
		if url == "" { return ToolResult{Success: false, Error: "SLACK_WEBHOOK_URL not set"} }
		if err := SendSlackMessage(url, message); err != nil {
			return ToolResult{Success: false, Error: err.Error()}
		}
	} else if channel == "sms" {
		// Twilio Simulation
		fmt.Printf("[SMS] Sending message: %s\n", message)
	} else {
		return ToolResult{Success: false, Error: "invalid channel"}
	}

	return ToolResult{
		Success: true,
		Data:    map[string]string{"status": "delivered", "channel": channel},
	}
}
