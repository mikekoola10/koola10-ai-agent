package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func agenticseekTool(payload map[string]interface{}) ToolResult {
	baseURL := os.Getenv("AGENTICSEEK_URL")
	if baseURL == "" {
		baseURL = "http://koola10-agenticseek.fly.dev"
	}

	action, ok := payload["action"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "Missing action in payload"}
	}

	switch action {
	case "browse":
		return handleAgenticBrowse(baseURL, payload)
	case "code":
		return handleAgenticCode(baseURL, payload)
	case "plan":
		return handleAgenticPlan(baseURL, payload)
	case "stripe_keys":
		return handleAgenticStripeKeys(baseURL, payload)
	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown action: %s", action)}
	}
}

func handleAgenticStripeKeys(baseURL string, payload map[string]interface{}) ToolResult {
	email := os.Getenv("STRIPE_LOGIN_EMAIL")
	password := os.Getenv("STRIPE_LOGIN_PASSWORD")

	if email == "" || password == "" {
		return ToolResult{Success: false, Error: "Missing STRIPE_LOGIN_EMAIL or STRIPE_LOGIN_PASSWORD environment variables"}
	}

	instruction := fmt.Sprintf("Navigate to Stripe, log in with email '%s' and password '%s'. If 2FA is required, return a screenshot. Otherwise, navigate to the API keys section and extract the Live Secret Key (sk_live_...) and Webhook Secret (whsec_...).", email, password)

	taskPayload := map[string]string{
		"instruction": instruction,
		"url":         "https://dashboard.stripe.com/login",
	}
	body, _ := json.Marshal(taskPayload)

	resp, err := http.Post(baseURL+"/task", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to call AgenticSeek: %v", err)}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to decode AgenticSeek response: %v", err)}
	}

	// Based on memory, if 2FA_REQUIRED is in status or result, it might return a screenshot
	status, _ := result["status"].(string)
	if status == "2FA_REQUIRED" {
		return ToolResult{
			Success: false,
			Error:   "2FA_REQUIRED",
			Data:    result,
		}
	}

	output, _ := result["result"].(string)
	if output == "" {
		output, _ = result["answer"].(string)
	}

	return ToolResult{
		Success: true,
		Output:  output,
		Data:    result,
	}
}

func handleAgenticCode(baseURL string, payload map[string]interface{}) ToolResult {
	language, _ := payload["language"].(string)
	code, _ := payload["code"].(string)

	if code == "" {
		return ToolResult{Success: false, Error: "Missing code for code action"}
	}

	taskPayload := map[string]string{
		"language": language,
		"code":     code,
	}
	body, _ := json.Marshal(taskPayload)

	// User prompt: "Sends to AgenticSeek's code execution endpoint"
	// Based on typical patterns, it might be /code
	resp, err := http.Post(baseURL+"/code", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to call AgenticSeek code endpoint: %v", err)}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	output, _ := result["output"].(string)
	return ToolResult{
		Success: true,
		Output:  output,
		Data:    result,
	}
}

func handleAgenticPlan(baseURL string, payload map[string]interface{}) ToolResult {
	goal, _ := payload["goal"].(string)

	if goal == "" {
		return ToolResult{Success: false, Error: "Missing goal for plan action"}
	}

	taskPayload := map[string]string{
		"goal": goal,
	}
	body, _ := json.Marshal(taskPayload)

	resp, err := http.Post(baseURL+"/plan", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to call AgenticSeek plan endpoint: %v", err)}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	output, _ := result["plan"].(string)
	return ToolResult{
		Success: true,
		Output:  output,
		Data:    result,
	}
}

func handleAgenticBrowse(baseURL string, payload map[string]interface{}) ToolResult {
	instruction, _ := payload["instruction"].(string)
	url, _ := payload["url"].(string)

	if instruction == "" {
		return ToolResult{Success: false, Error: "Missing instruction for browse action"}
	}

	// The user prompt specifically asked for /task
	taskPayload := map[string]string{
		"instruction": instruction,
		"url":         url,
	}
	body, _ := json.Marshal(taskPayload)

	resp, err := http.Post(baseURL+"/task", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to call AgenticSeek: %v", err)}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to decode AgenticSeek response: %v", err)}
	}

	output, _ := result["result"].(string)
	if output == "" {
		// fallback to answer if result is not there (based on api.py looking like it uses answer)
		output, _ = result["answer"].(string)
	}

	return ToolResult{
		Success: true,
		Output:  output,
		Data:    result,
	}
}

func init() {
	RegisterTool("agenticseek", agenticseekTool)
}
