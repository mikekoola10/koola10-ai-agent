package tools

import (
	"fmt"
	"log"
)

type AgentMailTool struct{}

func (a *AgentMailTool) SendEmail(to string, subject string, body string) error {
	log.Printf("[AgentMail] Sending to: %s | Subject: %s | Body: %s", to, subject, body)
	// Placeholder for actual email sending logic (e.g., SendGrid, Mailgun)
	return nil
}

func (a *AgentMailTool) SendTip(producer string, tip string) error {
	return a.SendEmail(producer, "Daily BeatSmith Production Tip", tip)
}

func (a *AgentMailTool) SendNudge(producer string, message string) error {
	return a.SendEmail(producer, "BeatSmith Nudge", message)
}

func init() {
	// Register with tools registry if applicable
	// This project uses a custom tool execution handler in main.go
}

func SendAlert(vertical string, message string) {
	am := &AgentMailTool{}
	am.SendEmail("admin@koola10.ai", fmt.Sprintf("[%s] System Alert", vertical), message)
}
