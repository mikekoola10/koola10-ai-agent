package main

import (
	"fmt"
	"log"
	"os"
	"koola10/sterling"
)

// Mock Ledger for local execution
type MockLedger struct{}
func (m *MockLedger) RecordCost(vertical, category string, amount float64, description string) {}
func (m *MockLedger) RecordRevenue(amount float64, source string) {
	fmt.Printf("[REVENUE] Recorded: %.2f from %s\n", amount, source)
}

func main() {
	deepseekAPIKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepseekAPIKey == "" {
		log.Fatal("DEEPSEEK_API_KEY is required")
	}

	ledger := &MockLedger{}
	vault := sterling.NewVaultClient()

	fmt.Println("--- Starting Local Emergency Swarm Execution ---")

	// 1. Bounty Hunter Swarm
	fmt.Println("\n[1/2] Launching Bounty Hunter Swarm...")
	bountyHunter := sterling.NewBountyHunter(ledger, vault, deepseekAPIKey)
	bountyHunter.RunDailyScan()

	// 2. Content Generator Swarm (Affiliate)
	fmt.Println("\n[2/2] Launching Content Generator (Affiliate) Swarm...")
	contentGen := sterling.NewContentGenerator(ledger, vault, deepseekAPIKey)
	contentGen.RunDailyContentCreation()

	fmt.Println("\n--- Emergency Swarm Execution Complete ---")
}
