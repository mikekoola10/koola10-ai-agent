package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type VaultEntry struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"` // income, expense, trade, profit
	Notes       string  `json:"notes"`
}

type VaultClient struct {
	apiURL string
}

func NewVaultClient(apiURL string) *VaultClient {
	return &VaultClient{apiURL: apiURL}
}

func (vc *VaultClient) AddEntry(entry VaultEntry) error {
	params := url.Values{}
	params.Add("action", "add")
	params.Add("description", entry.Description)
	params.Add("amount", fmt.Sprintf("%.2f", entry.Amount))
	params.Add("type", entry.Type)
	params.Add("notes", entry.Notes)

	fullURL := vc.apiURL + "?" + params.Encode()
	resp, err := http.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add entry, status: %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Status != "ok" {
		return fmt.Errorf("failed to add entry, status from API: %s", result.Status)
	}

	return nil
}
