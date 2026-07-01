package vault

import (
	"bytes"
	"encoding/json"
	"log"
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

func (vc *VaultClient) AddEntry(entry VaultEntry) error {
	log.Printf("[Vault] Adding entry: %+v", entry)
	if vc.apiURL == "" {
		return nil // Fallback if no API URL
	}

	body, _ := json.Marshal(entry)
	resp, err := http.Post(vc.apiURL+"/add", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
