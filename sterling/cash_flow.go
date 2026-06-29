package sterling

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"koola10/financial"
)

type Bill struct {
	ID           string    `json:"id"`
	Vendor       string    `json:"vendor"`
	Amount       float64   `json:"amount"`
	DueDate      time.Time `json:"due_date"`
	Paid         bool      `json:"paid"`
	PaymentTxID  string    `json:"payment_tx_id"`
	Recurring    bool      `json:"recurring"`
	IntervalDays int       `json:"interval_days"`
}

type CashFlow struct {
	ledger      *financial.EconomicLedger
	vaultClient interface{} // Placeholder for vault client if needed, or use ledger/audit
	bills       []Bill
	storagePath string
	mu          sync.RWMutex
}

func NewCashFlow(ledger *financial.EconomicLedger, storagePath string) *CashFlow {
	cf := &CashFlow{
		ledger:      ledger,
		storagePath: storagePath,
		bills:       []Bill{},
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
	if cf.bills == nil {
		cf.bills = []Bill{}
	}
}

func (cf *CashFlow) save() {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	data, _ := json.MarshalIndent(cf.bills, "", "  ")
	os.WriteFile(cf.storagePath, data, 0644)
}

func (cf *CashFlow) AddBill(vendor string, amount float64, dueDate time.Time, recurring bool, intervalDays int) {
	cf.mu.Lock()
	id := fmt.Sprintf("%x", time.Now().UnixNano())
	cf.bills = append(cf.bills, Bill{
		ID:           id,
		Vendor:       vendor,
		Amount:       amount,
		DueDate:      dueDate,
		Paid:         false,
		Recurring:    recurring,
		IntervalDays: intervalDays,
	})
	cf.mu.Unlock()
	cf.save()
}

func (cf *CashFlow) RunDailyPayer() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			cf.payDueBills()
			<-ticker.C
		}
	}()
	// Run once on start
	cf.payDueBills()
}

func (cf *CashFlow) GetLedger() *financial.EconomicLedger {
	return cf.ledger
}

func (cf *CashFlow) payDueBills() {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	now := time.Now()
	opsFund := cf.ledger.GetOperationsFund()

	log.Printf("[CashFlow] Checking due bills. Operations Fund: %.2f", opsFund)

	var nextBills []Bill

	for i, bill := range cf.bills {
		if !bill.Paid && (bill.DueDate.Before(now) || bill.DueDate.Equal(now)) {
			if opsFund >= bill.Amount {
				log.Printf("[CashFlow] Paying bill: %s to %s for %.2f", bill.ID, bill.Vendor, bill.Amount)

				txID, err := cf.ledger.RecordTransaction(
					fmt.Sprintf("Auto Bill Pay: %s", bill.Vendor),
					bill.Amount,
					"cost",
					fmt.Sprintf("Vendor: %s, BillID: %s", bill.Vendor, bill.ID),
				)
				if err == nil {
					cf.bills[i].Paid = true
					cf.bills[i].PaymentTxID = txID
					opsFund -= bill.Amount

					// Schedule next recurring bill
					if bill.Recurring {
						interval := bill.IntervalDays
						if interval == 0 {
							interval = 30
						}
						nextBills = append(nextBills, Bill{
							ID:           fmt.Sprintf("%x", time.Now().UnixNano()+int64(i)),
							Vendor:       bill.Vendor,
							Amount:       bill.Amount,
							DueDate:      bill.DueDate.AddDate(0, 0, interval),
							Paid:         false,
							Recurring:    true,
							IntervalDays: interval,
						})
					}
				} else {
					log.Printf("[CashFlow] Failed to record transaction: %v", err)
				}
			} else {
				log.Printf("[CashFlow] Insufficient funds to pay bill: %s to %s for %.2f", bill.ID, bill.Vendor, bill.Amount)
			}
		}
	}

	if len(nextBills) > 0 {
		cf.bills = append(cf.bills, nextBills...)
	}
	cf.save()
}
