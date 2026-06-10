package sterling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type PrivacyClient struct {
	APIKey string
}

type CardResponse struct {
	CardNumber string `json:"card_number"`
	ExpMonth   string `json:"exp_month"`
	ExpYear    string `json:"exp_year"`
	CVV        string `json:"cvv"`
	Hostname   string `json:"hostname,omitempty"`
	Memo       string `json:"memo"`
}

func NewPrivacyClient() *PrivacyClient {
	return &PrivacyClient{
		APIKey: os.Getenv("PRIVACY_API_KEY"),
	}
}

func (pc *PrivacyClient) CreateVirtualCard(memo string, spendLimitCents int) (*CardResponse, error) {
	if pc.APIKey == "" {
		return nil, fmt.Errorf("PRIVACY_API_KEY not set")
	}

	url := "https://api.privacy.com/v1/cards"
	payload := map[string]interface{}{
		"memo":                 memo,
		"spend_limit":          spendLimitCents,
		"spend_limit_duration": "MONTHLY",
		"type":                 "MERCHANT_LOCK",
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(pc.APIKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("privacy api error: %s", resp.Status)
	}

	var res struct {
		PAN      string `json:"pan"`
		ExpMonth int    `json:"exp_month"`
		ExpYear  int    `json:"exp_year"`
		CVV      string `json:"cvv"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &CardResponse{
		CardNumber: res.PAN,
		ExpMonth:   fmt.Sprintf("%02d", res.ExpMonth),
		ExpYear:    fmt.Sprintf("%d", res.ExpYear),
		CVV:        res.CVV,
		Memo:       memo,
	}, nil
}
