package financial

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type AgentCardClient struct {
	BaseURL string
	JWT     string
}

type CardResponse struct {
	ID          string `json:"id"`
	PAN         string `json:"pan,omitempty"`
	CVV         string `json:"cvv,omitempty"`
	ExpMonth    interface{} `json:"exp_month"` // Can be string or int
	ExpYear     interface{} `json:"exp_year"`  // Can be string or int
	Last4       string `json:"last4,omitempty"`
	Status      string `json:"status,omitempty"`
	AmountCents int    `json:"amountCents,omitempty"`
}

type CardDetails struct {
	PAN string `json:"pan"`
	CVV string `json:"cvv"`
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
	if ac.JWT == "" || ac.JWT == "YOUR_AGENTCARD_JWT" || os.Getenv("AGENTCARD_MOCK") == "true" {
		// Mock for development if JWT is missing, placeholder, or mock requested
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
		"amountCents":  int(amountLimit * 100), // Compatibility for both API versions
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

func (ac *AgentCardClient) GetCardDetails(cardID string) (*CardDetails, error) {
	req, _ := http.NewRequest("GET", ac.BaseURL+"/cards/"+cardID+"/details", nil)
	req.Header.Set("Authorization", "Bearer "+ac.JWT)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agentcard api error: %d", resp.StatusCode)
	}

	var details CardDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}
	return &details, nil
}

func (ac *AgentCardClient) ListCards() ([]CardResponse, error) {
	req, _ := http.NewRequest("GET", ac.BaseURL+"/cards", nil)
	req.Header.Set("Authorization", "Bearer "+ac.JWT)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agentcard api error: %d", resp.StatusCode)
	}

	var cards []CardResponse
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
		return nil, err
	}
	return cards, nil
}
