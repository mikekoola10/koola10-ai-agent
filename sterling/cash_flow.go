package sterling

import (
	"context"
	"log"
	"time"
	"koola10/financial"
	"koola10/vault"
)

type CashFlow struct {
	ledger *financial.FundManager
	vault  *vault.VaultClient
	bills  []financial.Bill
}

func NewCashFlow(ledger *financial.FundManager, vaultClient *vault.VaultClient) *CashFlow {
	return &CashFlow{ledger: ledger, vault: vaultClient}
}

func (cf *CashFlow) AddBill(vendor string, amount float64, dueDate time.Time) {
	cf.bills = append(cf.bills, financial.Bill{
		ID:      vendor + "_" + time.Now().Format("20060102150405"),
		Vendor:  vendor,
		Amount:  amount,
		DueDate: dueDate,
		Paid:    false,
	})
}

func (cf *CashFlow) RunDailyPayer(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run once on start
	cf.payPendingBills()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cf.payPendingBills()
		}
	}
}

func (cf *CashFlow) payPendingBills() {
	for i, b := range cf.bills {
		if !b.Paid && time.Now().After(b.DueDate) {
			ops := cf.ledger.GetOperationsFund() // 30% of total revenue
			if ops >= b.Amount {
				txID, _ := cf.ledger.RecordTransaction("Auto pay: "+b.Vendor, -b.Amount, "expense", "cash_flow")
				cf.vault.AddEntry(vault.VaultEntry{
					Description: "Auto bill: " + b.Vendor,
					Amount:      b.Amount,
					Type:        "expense",
					Notes:       "Paid by CashFlow",
				})
				cf.bills[i].Paid = true
				cf.bills[i].PaymentTxID = txID
				log.Printf("[Sterling] Paid %s: $%.2f", b.Vendor, b.Amount)
			} else {
				log.Printf("[Sterling] Insufficient ops fund for %s ($%.2f)", b.Vendor, b.Amount)
			}
		}
	}
}
