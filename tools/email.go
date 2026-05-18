package tools

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func init() {
	RegisterTool("email", emailTool)
}

func emailTool(payload map[string]interface{}) ToolResult {
	action, _ := payload["action"].(string)
	if action == "" {
		action = "send"
	}

	switch action {
	case "send":
		to, _ := payload["to"].(string)
		subject, _ := payload["subject"].(string)
		body, _ := payload["body"].(string)

		if to == "" || subject == "" || body == "" {
			return ToolResult{Success: false, Error: "Missing to, subject, or body"}
		}

		// Simulate sending email
		log.Printf("[EmailTool] Sending email to %s: %s", to, subject)

		// Log to a file for persistence/verification
		logDir := "./data/emails"
		os.MkdirAll(logDir, 0755)
		logFile := filepath.Join(logDir, "sent_emails.jsonl")

		entry := fmt.Sprintf(`{"timestamp":"%s", "to":"%s", "subject":"%s", "body":"%s"}`,
			time.Now().Format(time.RFC3339), to, subject, body)

		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return ToolResult{Success: false, Error: "Failed to open log file: " + err.Error()}
		}
		defer f.Close()

		if _, err := f.WriteString(entry + "\n"); err != nil {
			return ToolResult{Success: false, Error: "Failed to write to log file: " + err.Error()}
		}

		return ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Email sent to %s", to),
			Data: map[string]interface{}{
				"to":      to,
				"subject": subject,
				"status":  "sent",
			},
		}
	default:
		return ToolResult{Success: false, Error: "Unknown action: " + action}
	}
}
