package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"koola10/agents"
	"koola10/tools"
)

func RunOrchestration() {
	fmt.Println("🚀 Starting Revenue Acceleration and Swarm Scaling...")

	sm := agents.NewSwarmManager()
	sm.Factories["leadgen"] = agents.LeadGenFactory
	sm.Factories["content"] = agents.ContentFactory
	sm.Factories["trading"] = agents.TradingFactory

	// 1. REVENUE ACCELERATION
	fmt.Println("--- Revenue Acceleration ---")

	// Deploy Lead Gen Swarm
	err := sm.DeploySwarms("leadgen", 10)
	if err != nil {
		log.Fatalf("Failed to deploy LeadGen swarm: %v", err)
	}

	// Generate 50 prospects
	fmt.Println("Generating 50 nonprofit prospects...")
	res, err := sm.DispatchTask("leadgen", "generate_nonprofit_prospects")
	if err != nil {
		log.Fatalf("Failed to generate prospects: %v", err)
	}
	fmt.Printf("LeadGen Result: %v\n", res)

	// Read prospects
	prospectsData, err := os.ReadFile("./data/leads/prospects_nonprofit.json")
	if err != nil {
		log.Fatalf("Failed to read prospects: %v", err)
	}

	type Prospect struct {
		Name    string  `json:"name"`
		Mission string  `json:"mission"`
		Budget  float64 `json:"budget"`
		Email   string  `json:"email"`
	}
	var prospects []Prospect
	json.Unmarshal(prospectsData, &prospects)

	// Draft and Send Emails
	fmt.Println("Drafting and sending emails...")
	for i, p := range prospects {
		if i >= 50 { break }

		// Draft email
		prompt := fmt.Sprintf("Draft a personalized cold email to %s (Mission: %s). Reference their mission and recent grant activity. Include a clear CTA: 'Free one-week pilot of Spiral's AI grant discovery service.'", p.Name, p.Mission)
		draftRes := tools.RunTool("deepseek", map[string]interface{}{"prompt": prompt})

		if !draftRes.Success {
			log.Printf("Failed to draft email for %s: %s", p.Name, draftRes.Error)
			continue
		}

		emailBody := draftRes.Data.(map[string]interface{})["content"].(string)

		// Send first 10 emails
		if i < 10 {
			sendRes := tools.RunTool("email", map[string]interface{}{
				"to":      p.Email,
				"subject": fmt.Sprintf("Accelerating %s's Mission with AI", p.Name),
				"body":    emailBody,
			})
			if sendRes.Success {
				fmt.Printf("Email %d sent to %s\n", i+1, p.Email)
			} else {
				fmt.Printf("Failed to send email to %s: %s\n", p.Email, sendRes.Error)
			}
		}
	}

	// Generate Content with Solara
	fmt.Println("Generating Optimizr promotion content...")
	sm.DeploySwarms("content", 10)

	contentTasks := []string{
		"3 LinkedIn posts promoting Optimizr's $9/month image optimization API",
		"2 Reddit posts (r/nonprofit, r/grantwriting) promoting Optimizr's $9/month image optimization API",
		"1 Product Hunt comment promoting Optimizr's $9/month image optimization API",
	}

	for _, task := range contentTasks {
		res, err := sm.DispatchTask("content", task)
		if err != nil {
			fmt.Printf("Failed content task: %v\n", err)
		} else {
			fmt.Printf("Content Result: %v\n", res)
		}
	}

	// 2. SWARM SCALING
	fmt.Println("--- Swarm Scaling ---")

	// Based on requirements, we identify Lead Gen and Trading as top verticals
	verticalsToScale := []string{"leadgen", "trading"}

	for _, v := range verticalsToScale {
		fmt.Printf("Deploying 10 additional agents to %s vertical...\n", v)
		err := sm.DeploySwarms(v, 10)
		if err != nil {
			fmt.Printf("Failed to scale %s: %v\n", v, err)
		} else {
			fmt.Printf("Successfully scaled %s vertical.\n", v)
		}
	}

	fmt.Println("✅ Revenue Acceleration and Swarm Scaling tasks executed.")
}
