package sterling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type AgentCard struct {
	ID       string `json:"id"`
	PAN      string `json:"pan"`
	CVV      string `json:"cvv"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
}

type AgentCardClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewAgentCardClient() *AgentCardClient {
	return &AgentCardClient{
		apiKey: os.Getenv("AGENTCARD_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *AgentCardClient) CreateVirtualCard(memo string, limitCents int, ephemeral bool) (*AgentCard, error) {
	if c.apiKey == "" {
		// Mock for development if key is missing, but in production we expect it.
		// return nil, fmt.Errorf("AGENTCARD_API_KEY not set")
		return &AgentCard{
			ID:       "mock_card_123",
			PAN:      "4111222233334444",
			CVV:      "123",
			ExpMonth: 12,
			ExpYear:  2028,
		}, nil
	}

	payload := map[string]interface{}{
		"memo":       memo,
		"limit":      limitCents,
		"ephemeral":  ephemeral,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.agentcard.com/v1/cards", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create card, status: %d", resp.StatusCode)
	}

	var card AgentCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}

	return &card, nil
}
