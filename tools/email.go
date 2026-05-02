package tools

import (
	"fmt"
	"net/smtp"
	"os"
)

func emailTool(payload map[string]interface{}) ToolResult {
	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action"}
	}

	switch action {
	case "send":
		return sendEmail(payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func sendEmail(payload map[string]interface{}) ToolResult {
	to, _ := payload["to"].(string)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)

	if to == "" || subject == "" || body == "" {
		return ToolResult{Success: false, Error: "Missing to, subject, or body"}
	}

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	var status string
	if host == "" {
		fmt.Printf("MOCK EMAIL SENT to %s: %s\n", to, subject)
		status = "Email sent successfully (mocked)"
	} else {
		auth := smtp.PlainAuth("", user, pass, host)
		msg := []byte("To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n")
		err := smtp.SendMail(host+":"+port, auth, user, []string{to}, msg)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("Failed to send email: %v", err)}
		}
		status = "Email sent successfully via SMTP"
	}

	return ToolResult{Success: true, Output: status}
}

func init() {
	RegisterTool("email", emailTool)
}
