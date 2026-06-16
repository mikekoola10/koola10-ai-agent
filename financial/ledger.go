package financial

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type GlobalTransaction struct {
	Timestamp   string  `json:"timestamp"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Vertical    string  `json:"vertical,omitempty"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Notes       string  `json:"notes,omitempty"`
	TxID        string  `json:"tx_id,omitempty"`
}

type EconomicLedger struct {
	Balance      float64             `json:"balance"`
	TotalCosts   float64             `json:"total_costs"`
	TotalRevenue float64             `json:"total_revenue"`
	Transactions []GlobalTransaction `json:"transactions"`
	mu           sync.RWMutex
	storagePath  string
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
		storagePath:  path,
		Transactions: []GlobalTransaction{},
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
	data, _ := json.Marshal(l)
	os.WriteFile(l.storagePath, data, 0644)
}

func (l *EconomicLedger) RecordCost(vertical, category string, amount float64, description string) {
	l.mu.Lock()
	l.Balance -= amount
	l.TotalCosts += amount
	l.Transactions = append(l.Transactions, GlobalTransaction{
		Timestamp:   time.Now().Format(time.RFC3339),
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
	l.Transactions = append(l.Transactions, GlobalTransaction{
		Timestamp:   time.Now().Format(time.RFC3339),
		Type:        "revenue",
		Category:    "revenue_split",
		Vertical:    vertical,
		Amount:      amount,
		Description: "Revenue: " + source,
	})
	l.mu.Unlock()
	l.Save()
	if l.AuditLogger != nil {
		l.AuditLogger("economic_revenue_logged", map[string]interface{}{"amount": amount, "source": source, "vertical": vertical})
	}
}

func (l *EconomicLedger) GetBalance() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Balance
}

func (l *EconomicLedger) GetTotalRevenue() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.TotalRevenue
}

func (l *EconomicLedger) GetOperationsFund() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.TotalRevenue * 0.30
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

func (l *EconomicLedger) GetTransactions() []GlobalTransaction {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Transactions
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
	eval := EconomicEvaluation{
		Decision:      "allow",
		EstimatedCost: estimatedCost,
		ProjectedROI:  roi,
	}
	if roi < roiThreshold {
		eval.Decision = "warn"
		eval.Reason = "low_projected_roi"
	}
	l.mu.RLock()
	balance := l.Balance
	l.mu.RUnlock()
	if balance < estimatedCost {
		eval.Decision = "block"
		eval.Reason = "insufficient_funds"
	}
	return eval
}
