package sterling

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type VaultEntry struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Notes       string  `json:"notes"`
}

type VaultClient struct {
	apiURL string
}

func NewVaultClient() *VaultClient {
	return &VaultClient{
		apiURL: os.Getenv("VAULT_API_URL"),
	}
}

func (v *VaultClient) AddEntry(entry VaultEntry) error {
	if v.apiURL == "" {
		return nil
	}
	jsonData, _ := json.Marshal(entry)
	resp, err := http.Post(v.apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type Ledger interface {
	RecordRevenue(amount float64, source string)
	RecordCost(vertical, category string, amount float64, description string)
}
