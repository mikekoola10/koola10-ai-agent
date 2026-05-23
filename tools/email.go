package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func init() {
	RegisterTool("email", emailTool)
}

func emailTool(payload map[string]interface{}) ToolResult {
	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)

	if to == "" || subject == "" || body == "" {
		return ToolResult{Success: false, Error: "missing to, subject, or body"}
	}

	emailEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"to":        to,
		"subject":   subject,
		"body":      body,
	}

	logPath := "data/emails/sent_emails.jsonl"
	if _, err := os.Stat("/data"); err == nil {
		logPath = "/data/emails/sent_emails.jsonl"
	}

	os.MkdirAll(filepath.Dir(logPath), 0755)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return ToolResult{Success: false, Error: "failed to open email log"}
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(emailEntry); err != nil {
		return ToolResult{Success: false, Error: "failed to log email"}
	}

	return ToolResult{Success: true, Output: fmt.Sprintf("Email sent to %s", to)}
}
