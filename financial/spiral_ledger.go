package financial

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"koola10/tools"
)

type SpiralLedger struct {
	Balance      float64       `json:"balance"`
	TotalRevenue float64       `json:"total_revenue"`
	TotalSpent   float64       `json:"total_spent"`
	Transactions []Transaction `json:"transactions"`
	storagePath  string
	mu           sync.RWMutex
}

func NewSpiralLedger(path string) *SpiralLedger {
	sl := &SpiralLedger{
		storagePath:  path,
		Transactions: []Transaction{},
	}
	sl.load()
	return sl
}

func (sl *SpiralLedger) load() {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	data, err := os.ReadFile(sl.storagePath)
	if err == nil {
		json.Unmarshal(data, sl)
	}
}

func (sl *SpiralLedger) save() {
	data, err := json.MarshalIndent(sl, "", "  ")
	if err != nil {
		tools.LogStructured("ERROR", "internal", "spiral_ledger", "Failed to marshal ledger", map[string]interface{}{"error": err.Error()})
		return
	}
	err = os.WriteFile(sl.storagePath, data, 0644)
	if err != nil {
		tools.LogStructured("CRITICAL", "internal", "spiral_ledger", "Failed to save ledger to file", map[string]interface{}{"error": err.Error(), "path": sl.storagePath})
	}
}

func (sl *SpiralLedger) RecordRevenue(amount float64, source string) {
	sl.mu.Lock()
	oldBalance := sl.Balance
	sl.Balance += amount
	sl.TotalRevenue += amount
	sl.Transactions = append(sl.Transactions, Transaction{
		Timestamp:   time.Now(),
		Amount:      amount,
		Source:      source,
		Type:        "revenue",
		Description: fmt.Sprintf("Spiral Revenue from %s", source),
	})
	sl.save()
	sl.mu.Unlock()

	sl.CheckMilestones(oldBalance, sl.Balance)
}

func (sl *SpiralLedger) CheckMilestones(old, new float64) {
	milestones := []float64{100, 1000, 5000, 10000}
	for _, m := range milestones {
		if old < m && new >= m {
			sl.SendMilestoneAlert(m)
		}
	}
}

func (sl *SpiralLedger) SendMilestoneAlert(amount float64) {
	tools.RunTool("agentmail", map[string]interface{}{
		"to":      "mikekoola10@agentmail.to",
		"subject": fmt.Sprintf("Spiral Milestone Reached: $%.2f", amount),
		"body":    fmt.Sprintf("Congratulations! The Spiral Ecosystem has reached a new balance milestone: $%.2f. Current balance: $%.2f", amount, sl.Balance),
	})
}

func (sl *SpiralLedger) GetStatus() map[string]interface{} {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return map[string]interface{}{
		"balance":       sl.Balance,
		"total_revenue": sl.TotalRevenue,
		"total_spent":   sl.TotalSpent,
	}
}
