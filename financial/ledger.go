package financial

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Transaction struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Category    string    `json:"category,omitempty"`
	Vertical    string    `json:"vertical,omitempty"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Notes       string    `json:"notes,omitempty"`
	TxID        string    `json:"tx_id,omitempty"`
}

type EconomicLedger struct {
	Balance      float64       `json:"balance"`
	TotalCosts   float64       `json:"total_costs"`
	TotalRevenue float64       `json:"total_revenue"`
	Transactions []Transaction `json:"transactions"`
	mu           sync.RWMutex
	ledgerPath   string
	AuditLogger  func(action string, details map[string]interface{})
}

type EconomicSummary struct {
	Balance      float64 `json:"balance"`
	TotalCosts   float64 `json:"total_costs"`
	TotalRevenue float64 `json:"total_revenue"`
	ROI          float64 `json:"roi"`
}

type EconomicEvaluation struct {
	Decision      string  `json:"decision"`
	EstimatedCost float64 `json:"estimated_cost"`
	ProjectedROI  float64 `json:"projected_roi"`
	Reason        string  `json:"reason"`
}

func NewEconomicLedger(path string) *EconomicLedger {
	l := &EconomicLedger{
		Balance:      100.0,
		ledgerPath:   path,
		Transactions: []Transaction{},
	}
	l.Load()
	return l
}

func (l *EconomicLedger) Save() {
	l.mu.RLock()
	defer l.mu.RUnlock()
	data, _ := json.Marshal(l)
	os.WriteFile(l.ledgerPath, data, 0644)
}

func (l *EconomicLedger) Load() {
	l.mu.Lock()
	defer l.mu.Unlock()
	data, err := os.ReadFile(l.ledgerPath)
	if err == nil {
		json.Unmarshal(data, l)
	}
	if l.Transactions == nil {
		l.Transactions = []Transaction{}
	}
}

func (l *EconomicLedger) RecordCost(vertical, category string, amount float64, description string) {
	l.mu.Lock()
	l.Balance -= amount
	l.TotalCosts += amount
	l.Transactions = append(l.Transactions, Transaction{
		ID:          "tx_" + time.Now().Format("20060102150405"),
		Timestamp:   time.Now(),
		Type:        "cost",
		Category:    category,
		Vertical:    vertical,
		Amount:      amount,
		Description: description,
	})
	l.mu.Unlock()
	l.Save()
}

func (l *EconomicLedger) GetSummary() EconomicSummary {
	l.mu.RLock()
	defer l.mu.RUnlock()
	roi := 0.0
	if l.TotalCosts > 0 {
		roi = l.TotalRevenue / l.TotalCosts
	}
	return EconomicSummary{
		Balance:      l.Balance,
		TotalCosts:   l.TotalCosts,
		TotalRevenue: l.TotalRevenue,
		ROI:          roi,
	}
}

func (l *EconomicLedger) GetTransactions() []Transaction {
	l.mu.RLock()
	defer l.mu.RUnlock()
	txs := make([]Transaction, len(l.Transactions))
	copy(txs, l.Transactions)
	return txs
}

func (l *EconomicLedger) RecordRevenue(amount float64, source string) {
	l.RecordRevenueWithVertical("", amount, source)
}

func (l *EconomicLedger) RecordRevenueWithVertical(vertical string, amount float64, source string) {
	l.mu.Lock()
	l.Balance += amount
	l.TotalRevenue += amount
	l.Transactions = append(l.Transactions, Transaction{
		ID:          "tx_" + time.Now().Format("20060102150405"),
		Timestamp:   time.Now(),
		Type:        "revenue",
		Category:    "revenue_split",
		Vertical:    vertical,
		Amount:      amount,
		Description: "Revenue: " + source,
	})
	l.mu.Unlock()
	l.Save()
}

func (l *EconomicLedger) RecordTransaction(description string, amount float64, txType string, notes string) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	txID := "tx_" + time.Now().Format("20060102150405.000000")
	l.Transactions = append(l.Transactions, Transaction{
		ID:          txID,
		Timestamp:   time.Now(),
		Type:        txType,
		Amount:      amount,
		Description: description,
		Notes:       notes,
		TxID:        txID,
	})
	l.Balance += amount
	if amount > 0 {
		l.TotalRevenue += amount
	} else {
		l.TotalCosts -= amount
	}
	l.Save()
	return txID, nil
}

func (l *EconomicLedger) GetOperationsFund() float64 {
	return l.GetTotalRevenue() * 0.3
}

func (l *EconomicLedger) GetTotalRevenue() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var total float64
	for _, tx := range l.Transactions {
		if tx.Type == "income" || tx.Type == "profit" || tx.Type == "trade" || tx.Type == "revenue" {
			total += tx.Amount
		}
	}
	return total
}

func (l *EconomicLedger) GetBalance() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Balance
}

func (l *EconomicLedger) GetTotalCosts() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.TotalCosts
}

func (l *EconomicLedger) GetBills() []interface{} {
	return []interface{}{}
}
