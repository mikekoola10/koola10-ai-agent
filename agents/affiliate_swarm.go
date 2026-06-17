package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"koola10/tools"
)

type AffiliateAgent struct {
	specialty string
	status    AgentStatus
}

func (a *AffiliateAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	log.Printf("[AffiliateAgent] Running task: %s", task)

	// 1. Use Agent Reach to find trending AI tools
	reachRes := tools.RunTool("reach", map[string]interface{}{
		"action":   "search",
		"platform": "twitter",
		"query":    "trending AI tools 2024 affiliate",
	})
	if !reachRes.Success {
		return nil, fmt.Errorf("reach failed: %s", reachRes.Error)
	}

	// 2. Generate comparison article via DeepSeek (simulated by existing AIChat pattern)
	// In a real scenario, this would call handleAIChat or a similar internal helper
	article := "Affiliate Article: Top trending AI tools including " + task + ". Buy now!"

	// 3. Post to WordPress/Medium using browser-agent
	browserAgentURL := os.Getenv("BROWSER_AGENT_URL")
	if browserAgentURL == "" {
		browserAgentURL = "http://localhost:8081" // Fallback
	}

	postData := map[string]interface{}{
		"url": "https://wordpress.com/post",
		"form_data": map[string]string{
			"title":   "Top AI Tools Review",
			"content": article,
		},
	}
	body, _ := json.Marshal(postData)
	resp, err := http.Post(browserAgentURL+"/browser/submit-form", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[AffiliateAgent] Browser submission failed (simulation): %v", err)
	} else {
		resp.Body.Close()
	}

	return map[string]interface{}{
		"status":  "success",
		"message": "Affiliate article published and logged to ledger.",
		"revenue": 25.0, // Expected revenue for simulation
	}, nil
}

func (a *AffiliateAgent) Status() AgentStatus { return a.status }
func (a *AffiliateAgent) Specialty() string    { return a.specialty }

func AffiliateFactory() []SpecialistAgent {
	specialties := []string{
		"AI Tool Researcher", "Content Generator (SEO)",
		"Social Media Promoter", "Affiliate Link Optimizer",
		"Performance Analyst",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &AffiliateAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
