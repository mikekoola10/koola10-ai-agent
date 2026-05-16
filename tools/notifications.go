package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func init() {
	RegisterTool("notifications", notificationTool)
}

func notificationTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)

	switch action {
	case "send_desktop_alert":
		title, _ := payload["title"].(string)
		body, _ := payload["body"].(string)
		if title == "" || body == "" {
			return ToolResult{Success: false, Error: "title and body are required"}
		}

		// In a real scenario, this would broadcast an SSE event.
		// Since we're in a tool, we'll return the event data so the caller (main.go)
		// can broadcast it, OR we call an internal broadcast function if available.
		// For this implementation, we'll assume the goal is to trigger the dashboard's
		// notification system via the existing SSE mechanism.

		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Desktop alert sent: %s", title),
			Data: map[string]interface{}{
				"type": "notification",
				"title": title,
				"body":  body,
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}

	case "send_tts_prompt":
		text, _ := payload["text"].(string)
		if text == "" {
			return ToolResult{Success: false, Error: "text is required"}
		}

		// Call external TTS service
		ttsURL := "https://koola10-tts.fly.dev/speak"
		ttsReq := map[string]string{"text": text}
		jsonData, _ := json.Marshal(ttsReq)

		resp, err := http.Post(ttsURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("TTS service call failed: %v", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return ToolResult{Success: false, Error: fmt.Sprintf("TTS service returned status: %d", resp.StatusCode)}
		}

		return ToolResult{Success: true, Output: "TTS prompt sent"}

	default:
		return ToolResult{Success: false, Error: "Invalid action. Supported: send_desktop_alert, send_tts_prompt"}
	}
}
