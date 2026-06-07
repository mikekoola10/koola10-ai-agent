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

	// Mock response for demonstration when using the placeholder key
	if apiKey == "your_key" {
		action, _ := payload["action"].(string)
		if action == "create_agent_card" {
			merchant, _ := payload["merchant"].(string)
			if merchant == "" {
				merchant = "MockMerchant"
			}
			return ToolResult{
				Success: true,
				Output:  fmt.Sprintf("Card Created: %s\nPAN: 4111-1111-1111-1111\nExpiry: 01/2028\nCVV: 123", merchant),
				Data:    map[string]interface{}{"pan": "4111-1111-1111-1111", "exp_month": "01", "exp_year": "2028", "cvv": "123", "memo": merchant},
			}
		}
		return ToolResult{
			Success: true,
			Output:  "Card Created: Koola10 Subscription\nPAN: 4111-1111-1111-1111\nExpiry: 01/2028\nCVV: 123",
			Data:    map[string]interface{}{"pan": "4111-1111-1111-1111", "exp_month": "01", "exp_year": "2028", "cvv": "123"},
		}
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing or invalid 'action' in payload"}
	}

	switch action {
	case "create_subscription_card":
		return createSubscriptionCard(apiKey, payload)
	case "create_agent_card":
		return createAgentCard(apiKey, payload)
	case "create_card_browser":
		return createCardBrowser(payload)
	case "get_card_details":
		return getCardDetails(apiKey, payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		// Many APIs return month as a number. For exp_month, ensure 2 digits.
		return fmt.Sprintf("%02.0f", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func createCard(apiKey string, body map[string]interface{}) ToolResult {
	url := "https://api.privacy.com/v1/cards"

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to marshal request: %v", err)}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create request: %v", err)}
	}
	req.Header.Set("Authorization", "api-key "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to read response: %v", err)}
	}

	if resp.StatusCode >= 400 {
		return ToolResult{Success: false, Error: fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(respBody))}
	}

	var cardData map[string]interface{}
	if err := json.Unmarshal(respBody, &cardData); err != nil {
		return ToolResult{Success: false, Error: "Failed to parse response"}
	}

	pan := toString(cardData["pan"])
	expM := toString(cardData["exp_month"])
	expY := toString(cardData["exp_year"])
	cvv := toString(cardData["cvv"])
	memo := toString(cardData["memo"])

	// Update the map to ensure they are strings for downstream consumption
	cardData["pan"] = pan
	cardData["exp_month"] = expM
	cardData["exp_year"] = expY
	cardData["cvv"] = cvv

	output := fmt.Sprintf("Card Created: %s\nPAN: %s\nExpiry: %s/%s\nCVV: %s", memo, pan, expM, expY, cvv)
	return ToolResult{
		Success: true,
		Output:  output,
		Data:    cardData,
	}
}

func createAgentCard(apiKey string, payload map[string]interface{}) ToolResult {
	merchant, _ := payload["merchant"].(string)
	if merchant == "" {
		return ToolResult{Success: false, Error: "Missing merchant for agent card"}
	}

	limit, ok := payload["limit"].(float64)
	if !ok || limit <= 0 {
		limit = 5000 // Default $50.00
	}

	body := map[string]interface{}{
		"type":                 "MERCHANT_LOCK",
		"memo":                 merchant,
		"spend_limit":          int(limit),
		"spend_limit_duration": "TRANSACTION",
		"state":                "OPEN",
	}

	return createCard(apiKey, body)
}

func createCardBrowser(payload map[string]interface{}) ToolResult {
	browserAgentURL := os.Getenv("BROWSER_AGENT_URL")
	if browserAgentURL == "" {
		browserAgentURL = "http://localhost:8081"
	}

	email := os.Getenv("PRIVACY_EMAIL")
	password := os.Getenv("PRIVACY_PASSWORD")
	if email == "" || password == "" {
		return ToolResult{Success: false, Error: "PRIVACY_EMAIL or PRIVACY_PASSWORD not set"}
	}

	merchant, _ := payload["merchant"].(string)
	amountCents, _ := payload["limit"].(float64)
	memo, _ := payload["memo"].(string)
	if memo == "" {
		memo = merchant
	}

	reqBody := map[string]interface{}{
		"email":        email,
		"password":     password,
		"amount_cents": int(amountCents),
		"merchant":     merchant,
		"memo":         memo,
		"otp":          payload["otp"],
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post(browserAgentURL+"/browser/privacy/create-card", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return ToolResult{Success: false, Error: "Failed to connect to browser agent: " + err.Error()}
	}
	defer resp.Body.Close()

	var browserRes map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&browserRes); err != nil {
		return ToolResult{Success: false, Error: "Failed to parse browser agent response"}
	}

	if browserRes["status"] == "2FA_REQUIRED" {
		return ToolResult{
			Success: false,
			Error:   "2FA_REQUIRED",
			Data:    browserRes,
		}
	}

	if browserRes["status"] != "success" {
		return ToolResult{Success: false, Error: fmt.Sprintf("Browser automation failed: %v", browserRes["error"]), Data: browserRes}
	}

	card, _ := browserRes["card"].(map[string]interface{})
	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Card Created (via Browser): %s\nPAN: %s\nExpiry: %s/%s\nCVV: %s", memo, card["pan"], card["exp_month"], card["exp_year"], card["cvv"]),
		Data:    card,
	}
}

func createSubscriptionCard(apiKey string, payload map[string]interface{}) ToolResult {
	memo, _ := payload["memo"].(string)
	if memo == "" {
		memo = "Koola10 Subscription"
	}

	body := map[string]interface{}{
		"type":                 "MERCHANT_LOCK",
		"memo":                 memo,
		"spend_limit":          4000, // $40 in cents
		"spend_limit_duration": "MONTHLY",
		"state":                "OPEN",
	}

	return createCard(apiKey, body)
}

func getCardDetails(apiKey string, payload map[string]interface{}) ToolResult {
	cardToken, ok := payload["card_token"].(string)
	if !ok || cardToken == "" {
		return ToolResult{Success: false, Error: "Missing card_token"}
	}

	url := fmt.Sprintf("https://api.privacy.com/v1/cards/%s", cardToken)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create request: %v", err)}
	}
	req.Header.Set("Authorization", "api-key "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to read response: %v", err)}
	}

	if resp.StatusCode >= 400 {
		return ToolResult{Success: false, Error: fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(respBody))}
	}

	var cardData map[string]interface{}
	if err := json.Unmarshal(respBody, &cardData); err != nil {
		return ToolResult{Success: false, Error: "Failed to parse response"}
	}

	pan := toString(cardData["pan"])
	expM := toString(cardData["exp_month"])
	expY := toString(cardData["exp_year"])
	cvv := toString(cardData["cvv"])

	// Update the map to ensure they are strings for downstream consumption
	cardData["pan"] = pan
	cardData["exp_month"] = expM
	cardData["exp_year"] = expY
	cardData["cvv"] = cvv

	output := fmt.Sprintf("PAN: %s, Expiry: %s/%s, CVV: %s", pan, expM, expY, cvv)
	return ToolResult{
		Success: true,
		Output:  output,
		Data:    cardData,
	}
}

func init() {
	RegisterTool("privacy", privacyTool)
}
