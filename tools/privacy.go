package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func privacyTool(payload map[string]interface{}) ToolResult {
	apiKey := os.Getenv("PRIVACY_API_KEY")
	if apiKey == "" {
		return ToolResult{Success: false, Error: "PRIVACY_API_KEY not set"}
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	switch action {
	case "create_subscription_card":
		return createSubscriptionCard(apiKey, payload)
	case "get_card_details":
		return getCardDetails(apiKey, payload)
	case "list_active_cards":
		return listActiveCards(apiKey, payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func createSubscriptionCard(apiKey string, payload map[string]interface{}) ToolResult {
	url := "https://api.privacy.com/v1/cards"

	memo, _ := payload["memo"].(string)
	if memo == "" {
		memo = "Koola10 Subscription"
	}

	body := map[string]interface{}{
		"type":                 "MERCHANT_LOCKED",
		"memo":                 memo,
		"spend_limit":          4000, // $40 in cents
		"spend_limit_duration": "MONTHLY",
		"state":                "OPEN",
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "api-key "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ToolResult{Success: false, Error: fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(respBody))}
	}

	var cardData map[string]interface{}
	if err := json.Unmarshal(respBody, &cardData); err != nil {
		return ToolResult{Success: false, Error: "Failed to parse response"}
	}

	pan, _ := cardData["pan"].(string)
	expM, _ := cardData["exp_month"].(string)
	expY, _ := cardData["exp_year"].(string)
	cvv, _ := cardData["cvv"].(string)

	output := fmt.Sprintf("Card Created: %s\nPAN: %s\nExpiry: %s/%s\nCVV: %s", memo, pan, expM, expY, cvv)
	return ToolResult{
		Success: true,
		Output:  output,
		Data:    cardData,
	}
}

func getCardDetails(apiKey string, payload map[string]interface{}) ToolResult {
	cardToken, ok := payload["card_token"].(string)
	if !ok || cardToken == "" {
		return ToolResult{Success: false, Error: "Missing card_token"}
	}

	url := fmt.Sprintf("https://api.privacy.com/v1/cards/%s", cardToken)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "api-key "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ToolResult{Success: false, Error: fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(respBody))}
	}

	var cardData map[string]interface{}
	if err := json.Unmarshal(respBody, &cardData); err != nil {
		return ToolResult{Success: false, Error: "Failed to parse response"}
	}

	pan, _ := cardData["pan"].(string)
	expM, _ := cardData["exp_month"].(string)
	expY, _ := cardData["exp_year"].(string)
	cvv, _ := cardData["cvv"].(string)

	output := fmt.Sprintf("PAN: %s, Expiry: %s/%s, CVV: %s", pan, expM, expY, cvv)
	return ToolResult{
		Success: true,
		Output:  output,
		Data:    cardData,
	}
}

func listActiveCards(apiKey string, payload map[string]interface{}) ToolResult {
	url := "https://api.privacy.com/v1/cards"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "api-key "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ToolResult{Success: false, Error: fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(respBody))}
	}

	var cards []map[string]interface{}
	if err := json.Unmarshal(respBody, &cards); err != nil {
		return ToolResult{Success: false, Error: "Failed to parse response"}
	}

	var activeCards []map[string]interface{}
	for _, c := range cards {
		if state, ok := c["state"].(string); ok && state == "OPEN" {
			activeCards = append(activeCards, c)
		}
	}

	output := fmt.Sprintf("Found %d active cards", len(activeCards))
	return ToolResult{
		Success: true,
		Output:  output,
		Data:    activeCards,
	}
}

func init() {
	RegisterTool("privacy", privacyTool)
}
