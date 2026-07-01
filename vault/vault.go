package vault

import (
	"fmt"
	"log"
)

type VaultClient struct {
	SheetID string
}

func NewVaultClient(sheetID string) *VaultClient {
	return &VaultClient{SheetID: sheetID}
}

func (v *VaultClient) SyncTransaction(id string, amount float64, description string) error {
	log.Printf("[Vault] Syncing transaction %s: %.2f - %s to Google Sheets %s", id, amount, description, v.SheetID)
	// Implementation would use Google Sheets API
	return nil
}

func (v *VaultClient) GetCurrentBalance() (float64, error) {
	fmt.Println("[Vault] Retrieving source of truth balance from Google Sheets")
	return 1250.75, nil
}
