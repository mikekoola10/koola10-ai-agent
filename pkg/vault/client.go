package vault

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const VaultAPIBase = "https://script.google.com/macros/s/AKfycbxFFviECwZcWEZy9HLo2aEUlAqB-brL5MZFcn1OtTe8wYurw4G7AJltd5dHAE6bQRRg/exec"

type VaultEntry struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"` // income, expense, trade, profit
	Notes       string  `json:"notes"`
}

type VaultClient struct {
	httpClient *http.Client
}

func NewVaultClient() *VaultClient {
	return &VaultClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// AddEntry sends a transaction to the Google Sheet
func (c *VaultClient) AddEntry(entry VaultEntry) error {
	v := url.Values{}
	v.Set("action", "add")
	v.Set("description", entry.Description)
	v.Set("amount", fmt.Sprintf("%.2f", entry.Amount))
	v.Set("type", entry.Type)
	v.Set("notes", entry.Notes)

	fullURL := VaultAPIBase + "?" + v.Encode()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("vault returned status %d", resp.StatusCode)
	}
	return nil
}

// GetSummary retrieves totals for the day/week
func (c *VaultClient) GetSummary() (string, error) {
	v := url.Values{}
	v.Set("action", "summary")
	fullURL := VaultAPIBase + "?" + v.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("vault returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
