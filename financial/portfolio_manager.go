package financial

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type PortfolioManager struct {
	Balance      float64       `json:"balance"`
	SweepHistory []string      `json:"sweep_history"`
	LastSweepTs  time.Time     `json:"last_sweep_ts"`
	storagePath  string
	ledger       Ledger
	mu           sync.RWMutex
}

func NewPortfolioManager(path string, ledger Ledger) *PortfolioManager {
	pm := &PortfolioManager{
		storagePath:  path,
		ledger:       ledger,
		SweepHistory: []string{},
	}
	pm.load()
	return pm
}

func (pm *PortfolioManager) load() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	data, err := os.ReadFile(pm.storagePath)
	if err == nil {
		json.Unmarshal(data, pm)
	}
	if pm.SweepHistory == nil {
		pm.SweepHistory = []string{}
	}
}

func (pm *PortfolioManager) save() {
	data, _ := json.MarshalIndent(pm, "", "  ")
	os.WriteFile(pm.storagePath, data, 0644)
}

func (pm *PortfolioManager) SweepProfits(transactions interface{}) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, _ := json.Marshal(transactions)
	var txs []map[string]interface{}
	json.Unmarshal(data, &txs)

	var cryptoProfit float64
	for _, tx := range txs {
		vertical, _ := tx["vertical"].(string)
		txType, _ := tx["type"].(string)
		amount, _ := tx["amount"].(float64)
		tsStr, _ := tx["timestamp"].(string)
		ts, _ := time.Parse(time.RFC3339, tsStr)

		if ts.After(pm.LastSweepTs) && (vertical == "trading" || vertical == "trading_v2") {
			if txType == "revenue" {
				cryptoProfit += amount
			} else if txType == "cost" {
				cryptoProfit -= amount
			}
		}
	}

	if cryptoProfit > 0 {
		sweepAmount := cryptoProfit * 0.50
		pm.Balance += sweepAmount
		pm.LastSweepTs = time.Now()
		msg := fmt.Sprintf("Swept %.2f (50%% of %.2f crypto profit) into investment portfolio", sweepAmount, cryptoProfit)
		pm.SweepHistory = append(pm.SweepHistory, fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), msg))
		pm.save()
		fmt.Println(msg)
	}
}
