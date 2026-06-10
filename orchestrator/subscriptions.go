package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"koola10/sterling"
)

type ServiceInfo struct {
	Name        string  `json:"name"`
	URL         string  `json:"url"`
	MonthlyCost float64 `json:"monthly_cost"`
}

var AI_SERVICES = map[string]ServiceInfo{
	"chatgpt_plus":           {"ChatGPT Plus", "https://chat.openai.com", 20.0},
	"chatgpt_go":             {"ChatGPT Go", "https://chat.openai.com", 8.0},
	"gemini_advanced":        {"Gemini Advanced", "https://gemini.google.com", 20.0},
	"claude_pro":             {"Claude Pro", "https://claude.ai", 20.0},
	"grok_x":                 {"Grok (X Premium+)", "https://x.com", 40.0},
	"grok_standalone":        {"Grok Standalone", "https://x.ai", 30.0},
	"perplexity_pro":         {"Perplexity Pro", "https://perplexity.ai", 20.0},
	"midjourney_basic":       {"Midjourney Basic", "https://midjourney.com", 10.0},
	"midjourney_standard":    {"Midjourney Standard", "https://midjourney.com", 30.0},
	"suno_pro":               {"Suno Pro", "https://suno.ai", 10.0},
	"suno_premier":           {"Suno Premier", "https://suno.ai", 24.0},
	"adobe_firefly_standard": {"Adobe Firefly Standard", "https://firefly.adobe.com", 10.0},
	"adobe_firefly_pro":      {"Adobe Firefly Pro", "https://firefly.adobe.com", 20.0},
	"runway_gen3_pro":        {"Runway Gen-3 Pro", "https://runwayml.com", 35.0},
	"runway_unlimited":       {"Runway Unlimited", "https://runwayml.com", 76.0},
	"leonardo_ai_premium":    {"Leonardo.ai Premium", "https://leonardo.ai", 30.0},
}

type SubscriptionsManager struct {
	agentCardClient *sterling.AgentCardClient
	cashFlow        *sterling.CashFlow
	browserUrl      string
}

func NewSubscriptionsManager(acc *sterling.AgentCardClient, cf *sterling.CashFlow) *SubscriptionsManager {
	url := os.Getenv("BROWSER_AGENT_URL")
	if url == "" || url == "http://localhost:8081" {
		// Fallback to a sensible default if not set or explicitly set to localhost in prod
		url = "https://koola10-browser.fly.dev"
	}
	return &SubscriptionsManager{
		agentCardClient: acc,
		cashFlow:        cf,
		browserUrl:      url,
	}
}

func (sm *SubscriptionsManager) AutoSubscribe(serviceKey, userEmail string) error {
	info, ok := AI_SERVICES[serviceKey]
	if !ok {
		return fmt.Errorf("unknown service: %s", serviceKey)
	}

	log.Printf("[Orchestrator] Starting auto-subscribe for %s (%s)", info.Name, userEmail)

	// 1. Create virtual card
	limitCents := int(info.MonthlyCost * 100)
	card, err := sm.agentCardClient.CreateVirtualCard(fmt.Sprintf("AI Sub: %s", info.Name), limitCents, false)
	if err != nil {
		return fmt.Errorf("failed to create virtual card: %v", err)
	}

	password := "koola101" // Default password
	if userEmail == "" {
		userEmail = "mikekoola6@gmail.com"
	}

	// 2. Call browser agent
	signupReq := map[string]interface{}{
		"service":   serviceKey,
		"email":     userEmail,
		"password":  password,
		"card_info": map[string]interface{}{
			"pan":       card.PAN,
			"cvv":       card.CVV,
			"exp_month": card.ExpMonth,
			"exp_year":  card.ExpYear,
		},
	}

	jsonReq, _ := json.Marshal(signupReq)
	resp, err := http.Post(sm.browserUrl+"/signup_ai_service", "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		return fmt.Errorf("browser agent request failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status      string  `json:"status"`
		APIKey      string  `json:"api_key"`
		MonthlyCost float64 `json:"monthly_cost"`
		Message     string  `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode browser agent response: %v", err)
	}

	if result.Status == "manual_intervention_required" {
		log.Printf("[Orchestrator] %s requires manual intervention: %s", info.Name, result.Message)
		sm.cashFlow.GetLedger().RecordTransaction(fmt.Sprintf("Subscription Pending: %s", info.Name), 0, "cost", fmt.Sprintf("Requires manual: %s", result.Message))
		sm.cashFlow.AddBill(info.Name, info.MonthlyCost, time.Now().AddDate(0, 0, 30), true, 30, card.ID)
		return nil
	}

	if result.Status != "success" {
		return fmt.Errorf("signup failed: %s", result.Message)
	}

	log.Printf("[Orchestrator] Successfully subscribed to %s.", info.Name)

	if result.APIKey != "" {
		sm.cashFlow.GetLedger().RecordTransaction(fmt.Sprintf("API Key Secured: %s", info.Name), 0, "cost", fmt.Sprintf("Key: %s", result.APIKey))
	}

	// 3. Add bill to CashFlow (due in 30 days)
	sm.cashFlow.AddBill(info.Name, info.MonthlyCost, time.Now().AddDate(0, 0, 30), true, 30, card.ID)

	return nil
}

func (sm *SubscriptionsManager) SubscribeAll(userEmail string) {
	for key := range AI_SERVICES {
		err := sm.AutoSubscribe(key, userEmail)
		if err != nil {
			log.Printf("[Orchestrator] Failed to subscribe to %s: %v", key, err)
		}
		time.Sleep(30 * time.Second)
	}
}

func (sm *SubscriptionsManager) GetCashFlow() *sterling.CashFlow {
	return sm.cashFlow
}
