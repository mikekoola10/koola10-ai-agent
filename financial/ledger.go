package financial

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type GlobalTransaction struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // "revenue", "cost", "reinvestment"
	Vertical    string    `json:"vertical"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	AuditNote   string    `json:"audit_note"`
}

type EconomicLedger struct {
	Balance      float64             `json:"balance"`
	TotalCosts   float64             `json:"total_costs"`
	TotalRevenue float64             `json:"total_revenue"`
	Transactions []GlobalTransaction `json:"transactions"`
	mu           sync.RWMutex
	storagePath  string
}

func NewEconomicLedger(path string) *EconomicLedger {
	l := &EconomicLedger{
		storagePath: path,
	}
	l.Load()
	return l
}

func (l *EconomicLedger) Load() {
	l.mu.Lock()
	defer l.mu.Unlock()
	data, err := os.ReadFile(l.storagePath)
	if err == nil {
		json.Unmarshal(data, l)
	}
	if l.Transactions == nil {
		l.Transactions = []GlobalTransaction{}
	}
}

func (l *EconomicLedger) Save() {
	l.mu.RLock()
	defer l.mu.RUnlock()
	data, _ := json.MarshalIndent(l, "", "  ")
	os.WriteFile(l.storagePath, data, 0644)
}

func (l *EconomicLedger) RecordRevenue(vertical string, amount float64, source string) {
	l.mu.Lock()

	// 70/30 Gross Revenue Split
	// 70% stays in global ledger, 30% goes to operational fund (handled by FundManager)

	l.Balance += amount
	l.TotalRevenue += amount

	tx := GlobalTransaction{
		ID:          fmt.Sprintf("rev_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        "revenue",
		Vertical:    vertical,
		Category:    "revenue_split",
		Amount:      amount,
		Description: fmt.Sprintf("Revenue from %s", source),
		AuditNote:   "70% split applied",
	}
	l.Transactions = append(l.Transactions, tx)
	l.mu.Unlock()
	l.Save()
}

func (l *EconomicLedger) RecordCost(vertical, category string, amount float64, description string) {
	l.mu.Lock()
	l.Balance -= amount
	l.TotalCosts += amount

	tx := GlobalTransaction{
		ID:          fmt.Sprintf("cost_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Type:        "cost",
		Vertical:    vertical,
		Category:    category,
		Amount:      amount,
		Description: description,
	}
	l.Transactions = append(l.Transactions, tx)
	l.mu.Unlock()
	l.Save()
}

type EconomicEvaluation struct {
	Decision      string  `json:"decision"`
	EstimatedCost float64 `json:"estimated_cost"`
	ProjectedROI  float64 `json:"projected_roi"`
	Reason        string  `json:"reason"`
}

func (l *EconomicLedger) EvaluateAction(actionType string, estimatedCost float64) EconomicEvaluation {
	roiThreshold := 2.0
	projectedRevenue := 0.0
	if actionType == "grant_submit" {
		projectedRevenue = 500.0
	}

	roi := 0.0
	if estimatedCost > 0 {
		roi = projectedRevenue / estimatedCost
	}

	eval := EconomicEvaluation{"allow", estimatedCost, roi, ""}
	if roi < roiThreshold {
		eval.Decision = "warn"
		eval.Reason = "low_projected_roi"
	}

	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.Balance < estimatedCost {
		eval.Decision = "block"
		eval.Reason = "insufficient_funds"
	}
	return eval
}
