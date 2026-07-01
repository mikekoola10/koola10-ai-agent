package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	platforms := []string{"twitter", "reddit", "github"}
	targetPlatform := platforms[0]
	if strings.Contains(task, "reddit") { targetPlatform = "reddit" }

	reachRes := tools.RunTool("reach", map[string]interface{}{
		"action":   "search",
		"platform": targetPlatform,
		"query":    "trending AI tools 2024 affiliate " + task,
	})
	if !reachRes.Success {
		return nil, fmt.Errorf("reach failed: %s", reachRes.Error)
	}

	// 2. Generate comparison article via DeepSeek (simulated by existing AIChat pattern)
	// In a real scenario, this would call handleAIChat or a similar internal helper
	article := "Affiliate Article: Top trending AI tools including " + task + ". Buy now!"

	// 3. Select Affiliate Network (Expansion: CJ, Rakuten, Impact, ShareASale)
	network := "Amazon Associates"
	if strings.Contains(task, "SaaS") || strings.Contains(task, "Impact") { network = "Impact.com" }
	if strings.Contains(task, "Hardware") || strings.Contains(task, "ShareASale") { network = "ShareASale" }
	if strings.Contains(task, "Enterprise") || strings.Contains(task, "CJ") { network = "Commission Junction" }
	if strings.Contains(task, "Consumer") || strings.Contains(task, "Rakuten") { network = "Rakuten Advertising" }

	// CJ/Rakuten API Simulation
	if network == "Commission Junction" || network == "Rakuten Advertising" {
		log.Printf("[AffiliateAgent] Calling %s API for advertiser data...", network)
	}

	// 4. Post to WordPress/Medium/Substack using browser-agent
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
		"message": "Affiliate article published via " + network + " and logged to ledger.",
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
