package tools

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/issuing/card"
	"github.com/stripe/stripe-go/v76/issuing/cardholder"
)

func stripeIssuingTool(payload map[string]interface{}) ToolResult {
	apiKey := os.Getenv("STRIPE_API_KEY")
	if apiKey == "" {
		return ToolResult{Success: false, Error: "STRIPE_API_KEY not set"}
	}
	stripe.Key = apiKey

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	switch action {
	case "create_subscription_card":
		return createSubscriptionCard(payload)
	case "get_card_details":
		return getCardDetails(payload)
	case "destroy_card":
		return destroyCard(payload)
	case "list_active_cards":
		return listActiveCards(payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func findCardholderByName(name string) (*stripe.IssuingCardholder, error) {
	params := &stripe.IssuingCardholderListParams{}
	i := cardholder.List(params)
	for i.Next() {
		ch := i.IssuingCardholder()
		if ch.Name == name {
			return ch, nil
		}
	}
	if err := i.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("cardholder not found: %s", name)
}

func createSubscriptionCard(payload map[string]interface{}) ToolResult {
	ch, err := findCardholderByName("Koola10 Agent")
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	params := &stripe.IssuingCardParams{
		Cardholder: stripe.String(ch.ID),
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		Type:       stripe.String(string(stripe.IssuingCardTypeVirtual)),
		SpendingControls: &stripe.IssuingCardSpendingControlsParams{
			SpendingLimits: []*stripe.IssuingCardSpendingControlsSpendingLimitParams{
				{
					Amount:   stripe.Int64(4000), // $40.00
					Interval: stripe.String(string(stripe.IssuingCardholderSpendingControlsSpendingLimitIntervalMonthly)),
				},
			},
		},
	}
	params.AddExpand("number")
	params.AddExpand("cvc")

	c, err := card.New(params)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create card: %v", err)}
	}

	data := map[string]interface{}{
		"id":     c.ID,
		"number": c.Number,
		"cvc":    c.CVC,
		"exp_month": c.ExpMonth,
		"exp_year":  c.ExpYear,
		"brand":     c.Brand,
		"status":    c.Status,
	}

	output := fmt.Sprintf("Card Created: %s\nNumber: %s\nExp: %02d/%d\nCVC: %s", c.ID, c.Number, c.ExpMonth, c.ExpYear, c.CVC)

	return ToolResult{
		Success: true,
		Output:  output,
		Data:    data,
	}
}

func getCardDetails(payload map[string]interface{}) ToolResult {
	cardID, ok := payload["card_id"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing card_id"}
	}

	params := &stripe.IssuingCardParams{}
	params.AddExpand("number")
	params.AddExpand("cvc")

	c, err := card.Get(cardID, params)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to get card: %v", err)}
	}

	data := map[string]interface{}{
		"id":     c.ID,
		"number": c.Number,
		"cvc":    c.CVC,
		"exp_month": c.ExpMonth,
		"exp_year":  c.ExpYear,
		"brand":     c.Brand,
		"status":    c.Status,
	}

	output := fmt.Sprintf("Card: %s\nNumber: %s\nExp: %02d/%d\nCVC: %s", c.ID, c.Number, c.ExpMonth, c.ExpYear, c.CVC)

	return ToolResult{
		Success: true,
		Output:  output,
		Data:    data,
	}
}

func destroyCard(payload map[string]interface{}) ToolResult {
	cardID, ok := payload["card_id"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing card_id"}
	}

	params := &stripe.IssuingCardParams{
		Status: stripe.String(string(stripe.IssuingCardStatusCanceled)),
	}

	c, err := card.Update(cardID, params)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to destroy card: %v", err)}
	}

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Card %s destroyed (status: %s)", c.ID, c.Status),
		Data:    map[string]string{"id": c.ID, "status": string(c.Status)},
	}
}

func listActiveCards(payload map[string]interface{}) ToolResult {
	ch, err := findCardholderByName("Koola10 Agent")
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	params := &stripe.IssuingCardListParams{
		Cardholder: stripe.String(ch.ID),
		Status:     stripe.String(string(stripe.IssuingCardStatusActive)),
	}

	i := card.List(params)
	var cards []map[string]interface{}
	output := "Active Cards:\n"

	for i.Next() {
		c := i.IssuingCard()
		limitStr := "No limit"
		if len(c.SpendingControls.SpendingLimits) > 0 {
			limit := c.SpendingControls.SpendingLimits[0]
			limitStr = fmt.Sprintf("$%.2f/%s", float64(limit.Amount)/100.0, limit.Interval)
		}

		cards = append(cards, map[string]interface{}{
			"id":     c.ID,
			"last4":  c.Last4,
			"status": c.Status,
			"limit":  limitStr,
		})
		output += fmt.Sprintf("- %s (Last4: %s, Limit: %s)\n", c.ID, c.Last4, limitStr)
	}

	if err := i.Err(); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{
		Success: true,
		Output:  output,
		Data:    cards,
	}
}

func init() {
	RegisterTool("stripe_issuing", stripeIssuingTool)
}
