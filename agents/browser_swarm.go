package agents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type BrowserAgent struct {
	specialty string
	status    AgentStatus
	BaseURL   string
}

func (a *BrowserAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(task), &payload); err != nil {
		// If not JSON, treat as a direct extraction instruction
		payload = map[string]interface{}{
			"instruction": task,
			"url":         "https://www.google.com", // Default URL if not provided
		}
	}

	action, _ := payload["action"].(string)

	endpoint := "/browser/extract"
	if action == "stripe-live-keys" {
		endpoint = "/browser/stripe-live-keys"
	} else if action == "navigate" {
		endpoint = "/browser/navigate"
	} else if action == "fill-form" {
		endpoint = "/browser/fill-form"
	} else if action == "submit-form" {
		endpoint = "/browser/submit-form"
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(a.BaseURL+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		a.status = StatusError
		return nil, err
	}
	defer resp.Body.Close()

	var result interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (a *BrowserAgent) Status() AgentStatus { return a.status }
func (a *BrowserAgent) Specialty() string    { return a.specialty }

func BrowserFactory() []SpecialistAgent {
	url := os.Getenv("BROWSER_AGENT_URL")
	if url == "" {
		url = "https://koola10-browser.fly.dev"
	}
	specialties := []string{
		"Form Filler", "Data Extractor", "Interaction Specialist", "Visual Auditor", "Auth Handler",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BrowserAgent{
			specialty: s,
			status:    StatusIdle,
			BaseURL:   url,
		})
	}
	return agents
}
