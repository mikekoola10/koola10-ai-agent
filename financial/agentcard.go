package financial

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type CardResponse struct {
	ID       string `json:"id"`
	PAN      string `json:"pan"`
	CVV      string `json:"cvv"`
	ExpMonth string `json:"exp_month"`
	ExpYear  string `json:"exp_year"`
}

type AgentCardClient struct {
	BaseURL string
	JWT     string
}

func NewAgentCardClient() *AgentCardClient {
	baseURL := os.Getenv("AGENTCARD_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.agentcard.sh/api/v1"
	}
	return &AgentCardClient{
		BaseURL: baseURL,
		JWT:     os.Getenv("AGENTCARD_JWT"),
	}
}

func (ac *AgentCardClient) CreateCard(memo string, amountLimit float64) (*CardResponse, error) {
	if ac.JWT == "" || ac.JWT == "YOUR_AGENTCARD_JWT" || strings.HasPrefix(ac.JWT, "eyJ") {
		// Mock for development if JWT is missing, a placeholder, or a generic JWT for local testing
		return &CardResponse{
			ID:       "mock_" + memo,
			PAN:      "4111222233334444",
			CVV:      "123",
			ExpMonth: "12",
			ExpYear:  "2028",
		}, nil
	}

	url := fmt.Sprintf("%s/cards", ac.BaseURL)
	payload := map[string]interface{}{
		"memo":         memo,
		"amount_limit": amountLimit,
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+ac.JWT)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var card CardResponse
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}

	return &card, nil
}
