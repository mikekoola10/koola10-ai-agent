package agents

import (
	"fmt"
	"koola10/tools"
	"net/url"
)

type HealthPurchaserAgent struct {
	HealthAgent
}

func (a *HealthPurchaserAgent) Run(itemName string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// 1. Check if purchase needed (simulation: always yes for now if called)

	// 2. Create virtual card via AgentCard
	cardRes := tools.RunTool("agentcard", map[string]interface{}{
		"action":             "create_card",
		"memo":               "Health Purchase: " + itemName,
		"spend_limit_cents": 5000,
	})

	if !cardRes.Success {
		return nil, fmt.Errorf("failed to create card: %s", cardRes.Error)
	}

	card, _ := cardRes.Data.(*tools.CardResponse)

	// 3. Execute purchase via browser-agent
	// Simulation: navigating to a retailer and filling form
	purchaseRes := tools.RunTool("browser", map[string]interface{}{
		"action": "submit_form",
		"url":    "https://www.amazon.com/s?k=" + url.QueryEscape(itemName), // Encode item name
		"form_data": map[string]string{
			"card_number": card.PAN,
			"cvv":         card.CVV,
			"expiry":      card.ExpMonth + "/" + card.ExpYear,
		},
	})

	if !purchaseRes.Success {
		return nil, fmt.Errorf("purchase execution failed: %s", purchaseRes.Error)
	}

	// 4. Log in Economic Ledger (simulated callback or direct call if available)
	// For this task, we'll return the result which main.go will log.

	return map[string]interface{}{
		"item":    itemName,
		"status":  "purchased",
		"cost":    45.00, // Mock cost
		"details": purchaseRes.Data,
	}, nil
}
