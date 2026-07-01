package sterling

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type PrivacyClient struct {
	apiKey string
}

func NewPrivacyClient() *PrivacyClient {
	return &PrivacyClient{apiKey: os.Getenv("PRIVACY_API_KEY")}
}

type CardRequest struct {
	Memo        string `json:"memo"`
	SpendLimit  int    `json:"spend_limit"`
	LimitPeriod string `json:"limit_period"`
}

type CardResponse struct {
	Token    string `json:"token"`
	Last4    string `json:"last4"`
	Pan      string `json:"pan"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
	CVV      string `json:"cvv"`
}

func (pc *PrivacyClient) CreateVirtualCard(memo string, spendLimitCents int) (*CardResponse, error) {
	reqBody := CardRequest{
		Memo:        memo,
		SpendLimit:  spendLimitCents,
		LimitPeriod: "monthly",
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.privacy.com/v1/cards", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(pc.apiKey, "")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var card CardResponse
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}
	return &card, nil
}
