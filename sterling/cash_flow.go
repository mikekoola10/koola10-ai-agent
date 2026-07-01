package sterling

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
	"koola10/financial"
	"koola10/vault"
)

type CashFlow struct {
	ledger      financial.Ledger
	vault       *vault.VaultClient
	bills       []financial.Bill
	storagePath string
	mu          sync.RWMutex
}

func NewCashFlow(ledger financial.Ledger, vaultClient *vault.VaultClient) *CashFlow {
	path := "data/bills.json"
	if _, err := os.Stat("/data"); err == nil {
		path = "/data/bills.json"
	}

	cf := &CashFlow{
		ledger:      ledger,
		vault:       vaultClient,
		storagePath: path,
		bills:       []financial.Bill{},
	}
	cf.load()
	return cf
}

func (cf *CashFlow) load() {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	data, err := os.ReadFile(cf.storagePath)
	if err == nil {
		json.Unmarshal(data, &cf.bills)
	}
}

func (cf *CashFlow) save() {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	_ = os.MkdirAll(filepath.Dir(cf.storagePath), 0755)
	data, _ := json.MarshalIndent(cf.bills, "", "  ")
	_ = os.WriteFile(cf.storagePath, data, 0644)
}

func (cf *CashFlow) AddBill(vendor string, amount float64, dueDate time.Time) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	// Check if already exists to avoid duplicates
	for _, b := range cf.bills {
		if b.Vendor == vendor && b.Amount == amount && b.DueDate.Equal(dueDate) {
			return
		}
	}

	cf.bills = append(cf.bills, financial.Bill{
		ID:      vendor + "_" + time.Now().Format("20060102150405"),
		Vendor:  vendor,
		Amount:  amount,
		DueDate: dueDate,
		Paid:    false,
	})
	cf.save()
}

func (cf *CashFlow) RunDailyPayer(ctx context.Context) {
	// Initial check
	cf.checkAndPayBills()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cf.checkAndPayBills()
		}
	}
}

func (cf *CashFlow) checkAndPayBills() {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	changed := false
	for i, b := range cf.bills {
		if !b.Paid && time.Now().After(b.DueDate) {
			ops := cf.ledger.GetOperationsFund()
			if ops >= b.Amount {
				txID, err := cf.ledger.RecordTransaction("Auto pay: "+b.Vendor, -b.Amount, "expense", "cash_flow")
				if err != nil {
					log.Printf("[Sterling] Failed to record payment for %s: %v", b.Vendor, err)
					continue
				}
				_ = cf.vault.AddEntry(vault.VaultEntry{
					Description: "Auto bill: " + b.Vendor,
					Amount:      b.Amount,
					Type:        "expense",
					Notes:       "Paid by CashFlow",
				})
				cf.bills[i].Paid = true
				cf.bills[i].PaymentTxID = txID
				log.Printf("[Sterling] Paid %s: $%.2f", b.Vendor, b.Amount)
				changed = true
			} else {
				log.Printf("[Sterling] Insufficient ops fund for %s ($%.2f)", b.Vendor, b.Amount)
			}
		}
	}
	if changed {
		cf.save()
	}
}
