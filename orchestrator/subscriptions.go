package orchestrator

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"koola10/sterling"
)

type SubscriptionsManager struct {
	privacyClient *sterling.PrivacyClient
	cashFlow      *sterling.CashFlow
	browserUrl    string
}

func NewSubscriptionsManager(pc *sterling.PrivacyClient, cf *sterling.CashFlow) *SubscriptionsManager {
	url := os.Getenv("BROWSER_AGENT_URL")
	if url == "" {
		url = "https://koola10-browser-agent.fly.dev"
	}
	return &SubscriptionsManager{
		privacyClient: pc,
		cashFlow:      cf,
		browserUrl:    url,
	}
}

func (sm *SubscriptionsManager) AutoSubscribe(serviceName, userEmail string) error {
	log.Printf("[Orchestrator] Starting auto-subscribe for %s (%s)", serviceName, userEmail)

	// 1. Create virtual card
	// Using $20 limit for now as per instructions (2000 cents)
	card, err := sm.privacyClient.CreateVirtualCard(fmt.Sprintf("AI Sub: %s", serviceName), 2000)
	if err != nil {
		return fmt.Errorf("failed to create virtual card: %v", err)
	}

	password := generateRandomPassword()

	// 2. Call browser agent
	signupReq := map[string]interface{}{
		"service":   serviceName,
		"email":     userEmail,
		"password":  password,
		"card_info": card,
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

	if result.Status != "success" {
		return fmt.Errorf("signup failed: %s", result.Message)
	}

	log.Printf("[Orchestrator] Successfully subscribed to %s. Cost: %.2f", serviceName, result.MonthlyCost)

	// 3. Add bill to CashFlow (due next month)
	sm.cashFlow.AddBill(serviceName, result.MonthlyCost, time.Now().AddDate(0, 1, 0))

	return nil
}

func generateRandomPassword() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b) + "!A1"
}
