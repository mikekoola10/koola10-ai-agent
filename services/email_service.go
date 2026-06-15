package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type InboxResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type AgentMailClient struct {
	APIKey string
	BaseURL string
}

func NewAgentMailClient(apiKey string) *AgentMailClient {
	return &AgentMailClient{
		APIKey:  apiKey,
		BaseURL: "https://api.agentmail.to/v1",
	}
}

func (c *AgentMailClient) CreateInbox() (*InboxResponse, error) {
	req, _ := http.NewRequest("POST", c.BaseURL+"/inboxes", nil)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	var res InboxResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return &res, nil
}

func (c *AgentMailClient) SendEmail(to, subject, body string) error {
	mail := EmailRequest{To: to, Subject: subject, Body: body}
	data, _ := json.Marshal(mail)

	req, _ := http.NewRequest("POST", c.BaseURL+"/send", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("agentmail error: %d", resp.StatusCode)
	}
	return nil
}
