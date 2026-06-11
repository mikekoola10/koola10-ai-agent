package sterling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type VaultEntry struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Notes       string  `json:"notes,omitempty"`
}

type VaultClient struct {
	apiURL     string
	httpClient *http.Client
}

func NewVaultClient() *VaultClient {
	return &VaultClient{
		apiURL: os.Getenv("VAULT_API_URL"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (vc *VaultClient) AddEntry(entry VaultEntry) error {
	if vc.apiURL == "" {
		// return fmt.Errorf("VAULT_API_URL not set")
		return nil // Silent fail/mock
	}

	body, _ := json.Marshal(entry)
	resp, err := vc.httpClient.Post(vc.apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add vault entry, status: %d", resp.StatusCode)
	}

	return nil
}
