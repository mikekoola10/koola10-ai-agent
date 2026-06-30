package financial

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type AgentCardClient struct {
	JWT     string
	BaseURL string
}

type CardResponse struct {
	ID          string `json:"id"`
	Last4       string `json:"last4"`
	ExpMonth    int    `json:"exp_month"`
	ExpYear     int    `json:"exp_year"`
	Status      string `json:"status"`
	AmountCents int    `json:"amountCents"`
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
		JWT:     os.Getenv("AGENTCARD_JWT"),
		BaseURL: baseURL,
	}
}

func (c *AgentCardClient) CreateCard(amountCents int) (*CardResponse, error) {
	payload := map[string]interface{}{
		"amountCents": amountCents,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.BaseURL+"/cards", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+c.JWT)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errRes struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errRes)
		return nil, fmt.Errorf("agentcard api error (%d): %s", resp.StatusCode, errRes.Error)
	}

	var card CardResponse
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}
	return &card, nil
}

func (c *AgentCardClient) GetCardDetails(cardID string) (*CardDetails, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/cards/"+cardID+"/details", nil)
	req.Header.Set("Authorization", "Bearer "+c.JWT)

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

func (c *AgentCardClient) ListCards() ([]CardResponse, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/cards", nil)
	req.Header.Set("Authorization", "Bearer "+c.JWT)

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
