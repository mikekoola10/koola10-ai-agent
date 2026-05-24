package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type AuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func GetMachineAuthToken(agentName string) (string, error) {
	authURL := os.Getenv("MACHINE_AUTH_URL")
	if authURL == "" {
		authURL = "https://koola10-auth.fly.dev"
	}

	apiKey := os.Getenv("MACHINE_AUTH_ADMIN_KEY")

	payload := map[string]string{"name": agentName}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", authURL+"/api/tokens", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res AuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.AccessToken, nil
}
