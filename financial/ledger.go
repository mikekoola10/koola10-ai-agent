package financial

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Transaction struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // "cost", "revenue", "revenue_split", "expense", "reinvestment", "brief"
	Category    string    `json:"category,omitempty"`
	Vertical    string    `json:"vertical,omitempty"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Source      string    `json:"source,omitempty"`
}

type EconomicLedger struct {
	Balance           float64       `json:"balance"`
	TotalCosts        float64       `json:"total_costs"`
	TotalRevenue      float64       `json:"total_revenue"`
	OperationsSpent   float64       `json:"operations_spent"`
	Transactions      []Transaction `json:"transactions"`
	ledgerPath        string
	mu                sync.RWMutex
	AuditLogger       func(action string, details map[string]interface{}) `json:"-"`
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
		ledgerPath:   path,
		Transactions: []Transaction{},
	}
	l.Load()
	return l
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

func (l *EconomicLedger) Save() {
	l.mu.RLock()
	defer l.mu.RUnlock()
	data, _ := json.MarshalIndent(l, "", "  ")
	os.WriteFile(l.ledgerPath, data, 0644)
}

func (l *EconomicLedger) GetTotalRevenue() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.TotalRevenue
}

func (l *EconomicLedger) GetOperationsFund() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	// 30% of lifetime revenue minus what we spent operationally
	return (l.TotalRevenue * 0.3) - l.OperationsSpent
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

func (l *EconomicLedger) GetRevenueSplit() (total, ops, spend float64) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	total = l.TotalRevenue
	ops = (total * 0.3) - l.OperationsSpent
	spend = l.Balance - ops
	return
}

func (l *EconomicLedger) GetRecentTransactions(hours int) []Transaction {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var res []Transaction
	cutoff := time.Now().Add(time.Duration(-hours) * time.Hour)
	for _, t := range l.Transactions {
		if t.Timestamp.After(cutoff) {
			res = append(res, t)
		}
	}
	return res
}

func (l *EconomicLedger) RecordTransaction(desc string, amount float64, txType string, notes string) (string, error) {
	l.mu.Lock()
	id := generateHexID()
	t := Transaction{
		ID:          id,
		Timestamp:   time.Now(),
		Type:        txType,
		Amount:      amount,
		Description: desc,
	}
	l.Transactions = append(l.Transactions, t)

	if txType == "revenue" {
		l.Balance += amount
		l.TotalRevenue += amount
	} else if txType == "cost" {
		l.Balance -= amount
		l.TotalCosts += amount
		// If it's a bill pay, it's operational
		if len(desc) >= 13 && desc[:13] == "Auto Bill Pay" {
			l.OperationsSpent += amount
		}
	} else if txType != "brief" {
		l.Balance -= amount
		l.TotalCosts += amount
	}

	l.mu.Unlock()
	l.Save()
	if l.AuditLogger != nil {
		l.AuditLogger("transaction_recorded", map[string]interface{}{
			"id":          id,
			"description": desc,
			"amount":      amount,
			"type":        txType,
			"notes":       notes,
		})
	}
	return id, nil
}

func (l *EconomicLedger) RecordCost(vertical, category string, amount float64, description string) {
	l.mu.Lock()
	l.Balance -= amount
	l.TotalCosts += amount
	l.Transactions = append(l.Transactions, Transaction{
		ID:          generateHexID(),
		Timestamp:   time.Now(),
		Type:        "cost",
		Category:    category,
		Vertical:    vertical,
		Amount:      amount,
		Description: description,
	})
	l.mu.Unlock()
	l.Save()
	if l.AuditLogger != nil {
		l.AuditLogger("economic_cost_logged", map[string]interface{}{"amount": amount, "category": category, "vertical": vertical})
	}
}

func (l *EconomicLedger) RecordRevenue(amount float64, source string) {
	l.RecordRevenueWithVertical("", amount, source)
}

func (l *EconomicLedger) RecordRevenueWithVertical(vertical string, amount float64, source string) {
	l.mu.Lock()
	l.Balance += amount
	l.TotalRevenue += amount
	l.Transactions = append(l.Transactions, Transaction{
		ID:          generateHexID(),
		Timestamp:   time.Now(),
		Type:        "revenue",
		Category:    "revenue_split",
		Vertical:    vertical,
		Amount:      amount,
		Description: "Revenue: " + source,
		Source:      source,
	})
	l.mu.Unlock()
	l.Save()
	if l.AuditLogger != nil {
		l.AuditLogger("economic_revenue_logged", map[string]interface{}{"amount": amount, "source": source, "vertical": vertical})
	}
}

func (l *EconomicLedger) GetTransactions() []Transaction {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Transactions
}

func generateHexID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
