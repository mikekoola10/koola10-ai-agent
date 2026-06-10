package sterling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type AgentCardClient struct {
	APIKey string
}

type CardResponse struct {
	ID         string `json:"id"`
	PAN        string `json:"pan"`
	CVV        string `json:"cvv"`
	ExpMonth   string `json:"exp_month"`
	ExpYear    string `json:"exp_year"`
	Memo       string `json:"memo"`
}

func NewAgentCardClient() *AgentCardClient {
	return &AgentCardClient{
		APIKey: os.Getenv("AGENTCARD_API_KEY"),
	}
}

func (ac *AgentCardClient) CreateVirtualCard(memo string, spendLimitCents int, autoDestruct bool) (*CardResponse, error) {
	if ac.APIKey == "" {
		return nil, fmt.Errorf("AGENTCARD_API_KEY not set")
	}

	url := "https://api.agentcard.com/v1/cards"
	payload := map[string]interface{}{
		"memo":          memo,
		"spend_limit":   spendLimitCents,
		"auto_destruct": autoDestruct,
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+ac.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("agentcard api error: %s", resp.Status)
	}

	var res CardResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (ac *AgentCardClient) GetCardDetails(cardID string) (*CardResponse, error) {
	if ac.APIKey == "" {
		return nil, fmt.Errorf("AGENTCARD_API_KEY not set")
	}

	url := fmt.Sprintf("https://api.agentcard.com/v1/cards/%s", cardID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+ac.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agentcard api error: %s", resp.Status)
	}

	var res CardResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (ac *AgentCardClient) BlockCard(cardID string) error {
	if ac.APIKey == "" {
		return fmt.Errorf("AGENTCARD_API_KEY not set")
	}

	url := fmt.Sprintf("https://api.agentcard.com/v1/cards/%s/block", cardID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+ac.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agentcard api error: %s", resp.Status)
	}

	return nil
}
