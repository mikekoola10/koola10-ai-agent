package tools

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/subscription"
)

func stripeTool(payload map[string]interface{}) ToolResult {
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
	case "create_checkout_session":
		return createCheckoutSession(payload)
	case "get_subscription":
		return getSubscription(payload)
	case "cancel_subscription":
		return cancelSubscription(payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func createCheckoutSession(payload map[string]interface{}) ToolResult {
	priceID, _ := payload["price_id"].(string)
	customerEmail, _ := payload["customer_email"].(string)
	successURL, _ := payload["success_url"].(string)
	cancelURL, _ := payload["cancel_url"].(string)
	mode, _ := payload["mode"].(string) // "payment" or "subscription"

	if priceID == "" || successURL == "" || cancelURL == "" {
		return ToolResult{Success: false, Error: "Missing required parameters (price_id, success_url, cancel_url)"}
	}

	if mode == "" {
		mode = string(stripe.CheckoutSessionModeSubscription)
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:          stripe.String(mode),
		SuccessURL:    stripe.String(successURL),
		CancelURL:     stripe.String(cancelURL),
		CustomerEmail: stripe.String(customerEmail),
	}

	s, err := session.New(params)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create session: %v", err)}
	}

	return ToolResult{
		Success: true,
		Output:  s.URL,
		Data:    map[string]string{"url": s.URL, "id": s.ID},
	}
}

func getSubscription(payload map[string]interface{}) ToolResult {
	subID, _ := payload["subscription_id"].(string)
	if subID == "" {
		return ToolResult{Success: false, Error: "Missing subscription_id"}
	}

	s, err := subscription.Get(subID, nil)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to get subscription: %v", err)}
	}

	data := map[string]interface{}{
		"status":                 s.Status,
		"current_period_end":     s.CurrentPeriodEnd,
		"current_period_start":   s.CurrentPeriodStart,
		"cancel_at_period_end":   s.CancelAtPeriodEnd,
	}

	return ToolResult{
		Success: true,
		Output:  string(s.Status),
		Data:    data,
	}
}

func cancelSubscription(payload map[string]interface{}) ToolResult {
	subID, _ := payload["subscription_id"].(string)
	if subID == "" {
		return ToolResult{Success: false, Error: "Missing subscription_id"}
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	s, err := subscription.Update(subID, params)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to cancel subscription: %v", err)}
	}

	return ToolResult{
		Success: true,
		Output:  "Subscription set to cancel at period end",
		Data:    map[string]bool{"cancel_at_period_end": s.CancelAtPeriodEnd},
	}
}

func init() {
	RegisterTool("stripe", stripeTool)
}
