package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type VercelTokenRequest struct {
	Name string `json:"name"`
}

type VercelTokenResponse struct {
	Token struct {
		Token string `json:"token"`
	} `json:"token"`
	BearerToken string `json:"bearerToken"`
}

func vercelTokenTool(payload map[string]interface{}) ToolResult {
	vercelToken := os.Getenv("VERCEL_TOKEN")
	if vercelToken == "" {
		return ToolResult{Success: false, Error: "VERCEL_TOKEN not set"}
	}

	tokenName, ok := payload["token_name"].(string)
	if !ok || tokenName == "" {
		tokenName = "Jules Auto Token"
	}

	reqBody, _ := json.Marshal(VercelTokenRequest{Name: tokenName})
	req, err := http.NewRequest("POST", "https://api.vercel.com/v3/user/tokens", bytes.NewBuffer(reqBody))
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+vercelToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Vercel API request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ToolResult{Success: false, Error: fmt.Sprintf("Vercel API error (status %d): %s", resp.StatusCode, string(body))}
	}

	var res VercelTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to decode Vercel response: %v", err)}
	}

	// In newer Vercel API, it might just return 'token' or 'bearerToken'.
	// Based on the prompt, we return the bearer token.
	tokenValue := res.BearerToken
	if tokenValue == "" {
		tokenValue = res.Token.Token
	}

	return ToolResult{
		Success: true,
		Output:  "Successfully created Vercel token: " + tokenName,
		Data:    map[string]string{"bearer_token": tokenValue},
	}
}

func init() {
	RegisterTool("vercel_token", vercelTokenTool)
}
