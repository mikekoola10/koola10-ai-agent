package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type StackingTransaction struct {
	Timestamp   time.Time `json:"timestamp"`
	Amount      float64   `json:"amount"`
	Source      string    `json:"source"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
}

type StackingLedger struct {
	Balance      float64              `json:"balance"`
	TotalEarned  float64              `json:"total_earned"`
	TotalSpent   float64              `json:"total_spent"`
	Transactions []StackingTransaction `json:"transactions"`
	mu           sync.RWMutex
	path         string
}

func LoadStackingLedger(path string) *StackingLedger {
	sl := &StackingLedger{path: path}
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, sl)
	}
	return sl
}

func (sl *StackingLedger) Save() {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	data, _ := json.MarshalIndent(sl, "", "  ")
	os.WriteFile(sl.path, data, 0644)
}

func (sl *StackingLedger) RecordProfit(amount float64, source, description string) {
	sl.mu.Lock()
	sl.Balance += amount
	sl.TotalEarned += amount
	sl.Transactions = append(sl.Transactions, StackingTransaction{
		Timestamp:   time.Now(),
		Amount:      amount,
		Source:      source,
		Type:        "profit",
		Description: description,
	})
	sl.mu.Unlock()
	sl.Save()
}

type TradingPoolAgent struct {
	ledger *StackingLedger
}

func NewTradingPoolAgent(ledger *StackingLedger) *TradingPoolAgent {
	return &TradingPoolAgent{ledger: ledger}
}

func (a *TradingPoolAgent) RunMomentumStrategy() {
	// Simulate paper trading logic for Momentum strategy
	// 0.01 BTC, BTC/ETH 4-hour timeframe
	fmt.Println("[Trading Pool] Running Momentum strategy (0.01 BTC, BTC/ETH 4h)")

	// Simulated profit
	profit := 12.50 // Simulated $12.50 profit
	a.ledger.RecordProfit(profit, "trading_pool", "Momentum strategy (0.01 BTC, BTC/ETH 4h) paper trade")
}
