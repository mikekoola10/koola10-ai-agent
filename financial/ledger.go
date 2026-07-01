package financial

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type EconomicTransaction struct {
	Timestamp   string  `json:"timestamp"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Vertical    string  `json:"vertical,omitempty"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Settled     bool    `json:"settled"`
}

type EconomicLedger struct {
	Balance      float64               `json:"balance"`
	TotalCosts   float64               `json:"total_costs"`
	TotalRevenue float64               `json:"total_revenue"`
	Transactions []EconomicTransaction `json:"transactions"`
	Mu           sync.RWMutex
	ledgerPath   string
}

func NewEconomicLedger(path string) *EconomicLedger {
	l := &EconomicLedger{
		Balance:      100.0,
		ledgerPath:   path,
		Transactions: []EconomicTransaction{},
	}
	l.Load()
	return l
}

func (l *EconomicLedger) Save() {
	l.Mu.RLock()
	defer l.Mu.RUnlock()
	data, _ := json.Marshal(l)
	os.WriteFile(l.ledgerPath, data, 0644)
}

func (l *EconomicLedger) Load() {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	data, err := os.ReadFile(l.ledgerPath)
	if err == nil {
		json.Unmarshal(data, l)
	}
	if l.Transactions == nil {
		l.Transactions = []EconomicTransaction{}
	}
}

func (l *EconomicLedger) RecordCost(vertical, category string, amount float64, description string) {
	l.Mu.Lock()
	l.Balance -= amount
	l.TotalCosts += amount
	l.Transactions = append(l.Transactions, EconomicTransaction{
		Timestamp:   time.Now().Format(time.RFC3339),
		Type:        "cost",
		Category:    category,
		Vertical:    vertical,
		Amount:      amount,
		Description: description,
		Settled:     true,
	})
	l.Mu.Unlock()
	l.Save()
}

func (l *EconomicLedger) RecordRevenue(amount float64, source string) {
	l.RecordRevenueWithVertical("", amount, source, false) // Default to unsettled
}

func (l *EconomicLedger) RecordRevenueWithVertical(vertical string, amount float64, source string, settled bool) {
	l.Mu.Lock()
	l.Balance += amount
	l.TotalRevenue += amount
	l.Transactions = append(l.Transactions, EconomicTransaction{
		Timestamp:   time.Now().Format(time.RFC3339),
		Type:        "revenue",
		Category:    "revenue_split",
		Vertical:    vertical,
		Amount:      amount,
		Description: "Revenue: " + source,
		Settled:     settled,
	})
	l.Mu.Unlock()
	l.Save()
}

type Reconciliation struct {
	Timestamp      time.Time `json:"timestamp"`
	LedgerBalance  float64   `json:"ledger_balance"`
	ActualBalances map[string]float64 `json:"actual_balances"`
	Discrepancy    float64   `json:"discrepancy"`
	Status         string    `json:"status"` // "match", "discrepancy"
}

func (l *EconomicLedger) Reconcile(actual map[string]float64) Reconciliation {
	var totalActual float64
	for _, b := range actual {
		totalActual += b
	}

	discrepancy := l.Balance - totalActual
	status := "match"
	if discrepancy > 5.0 || discrepancy < -5.0 {
		status = "discrepancy"
	}

	return Reconciliation{
		Timestamp:      time.Now(),
		LedgerBalance:  l.Balance,
		ActualBalances: actual,
		Discrepancy:    discrepancy,
		Status:         status,
	}
}
