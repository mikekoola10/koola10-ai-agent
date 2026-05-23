package financial

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Transaction struct {
	Timestamp   time.Time `json:"timestamp"`
	Amount      float64   `json:"amount"`
	Source      string    `json:"source"`
	Type        string    `json:"type"` // "revenue", "expense", "reinvestment", "cold_storage_distribution"
	Description string    `json:"description"`
}

type StackingFundStatus struct {
	Balance             float64  `json:"balance"`
	TotalEarned         float64  `json:"total_earned"`
	TotalSpent          float64  `json:"total_spent"`
	ReinvestmentHistory []string `json:"reinvestment_history"`
}

type StackingFundManager struct {
	Balance             float64       `json:"balance"`
	TotalEarned         float64       `json:"total_earned"`
	TotalSpent          float64       `json:"total_spent"`
	ReinvestmentHistory []string      `json:"reinvestment_history"`
	Transactions        []Transaction `json:"transactions"`
	storagePath         string
	ledgerPath          string
	mu                  sync.RWMutex
}

func NewStackingFundManager(fundPath, ledgerPath string) *StackingFundManager {
	sfm := &StackingFundManager{
		storagePath:         fundPath,
		ledgerPath:          ledgerPath,
		ReinvestmentHistory: []string{},
		Transactions:        []Transaction{},
	}
	sfm.load()
	return sfm
}

func (sfm *StackingFundManager) load() {
	sfm.mu.Lock()
	defer sfm.mu.Unlock()
	data, err := os.ReadFile(sfm.storagePath)
	if err == nil {
		json.Unmarshal(data, sfm)
	}
	if sfm.ReinvestmentHistory == nil {
		sfm.ReinvestmentHistory = []string{}
	}
	if sfm.Transactions == nil {
		sfm.Transactions = []Transaction{}
	}
}

func (sfm *StackingFundManager) save() {
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(sfm.storagePath), 0755)

	data, _ := json.MarshalIndent(sfm, "", "  ")
	os.WriteFile(sfm.storagePath, data, 0644)

	// Also sync to a standalone ledger file for compliance
	ledgerData, _ := json.MarshalIndent(map[string]interface{}{
		"balance":      sfm.Balance,
		"total_earned": sfm.TotalEarned,
		"total_spent":  sfm.TotalSpent,
		"transactions": sfm.Transactions,
		"last_sync":    time.Now().Format(time.RFC3339),
	}, "", "  ")
	os.WriteFile(sfm.ledgerPath, ledgerData, 0644)
}

func (sfm *StackingFundManager) RecordRevenue(amount float64, source string, pillar string) {
	sfm.mu.Lock()
	defer sfm.mu.Unlock()

	sfm.Balance += amount
	sfm.TotalEarned += amount

	sfm.Transactions = append(sfm.Transactions, Transaction{
		Timestamp:   time.Now(),
		Amount:      amount,
		Source:      source,
		Type:        "revenue",
		Description: fmt.Sprintf("Revenue from pillar %s: %s", pillar, source),
	})

	// Check for reinvestment and profit distribution rules
	sfm.handleEconomics()
	sfm.save()
}

func (sfm *StackingFundManager) handleEconomics() {
	// Reinvestment engine: Auto-allocates 50% trading, 30% affiliate scaling, 20% reserve
	// This is called whenever revenue is recorded. For now, we'll log the allocation.
	// In a real scenario, this would trigger deployment of more agents or increased budgets.

	lastTx := sfm.Transactions[len(sfm.Transactions)-1]
	if lastTx.Type == "revenue" {
		tradingAlloc := lastTx.Amount * 0.50
		affiliateAlloc := lastTx.Amount * 0.30
		reserveAlloc := lastTx.Amount * 0.20

		msg := fmt.Sprintf("Auto-allocated revenue (%.2f): Trading: %.2f, Affiliate: %.2f, Reserve: %.2f",
			lastTx.Amount, tradingAlloc, affiliateAlloc, reserveAlloc)

		sfm.ReinvestmentHistory = append(sfm.ReinvestmentHistory, fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), msg))
	}

	// Profit distribution rule: If Stacking Fund reaches $50K+ in accumulated profits,
	// distribute 10% to a long-term cold storage wallet (BTC/ETH)
	if sfm.TotalEarned >= 50000.0 {
		// Calculate how much should have been distributed
		targetDistribution := sfm.TotalEarned * 0.10

		// Find how much was already distributed
		alreadyDistributed := 0.0
		for _, tx := range sfm.Transactions {
			if tx.Type == "cold_storage_distribution" {
				alreadyDistributed += tx.Amount
			}
		}

		pendingDistribution := targetDistribution - alreadyDistributed
		if pendingDistribution > 0 && sfm.Balance >= pendingDistribution {
			sfm.Balance -= pendingDistribution
			sfm.TotalSpent += pendingDistribution
			sfm.Transactions = append(sfm.Transactions, Transaction{
				Timestamp:   time.Now(),
				Amount:      pendingDistribution,
				Source:      "internal",
				Type:        "cold_storage_distribution",
				Description: fmt.Sprintf("10%% distribution of accumulated profits (%.2f) to cold storage", sfm.TotalEarned),
			})

			msg := fmt.Sprintf("Distributed %.2f to cold storage (10%% of %.2f total earned)", pendingDistribution, sfm.TotalEarned)
			sfm.ReinvestmentHistory = append(sfm.ReinvestmentHistory, fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), msg))
		}
	}
}

func (sfm *StackingFundManager) GetStatus() StackingFundStatus {
	sfm.mu.RLock()
	defer sfm.mu.RUnlock()
	return StackingFundStatus{
		Balance:             sfm.Balance,
		TotalEarned:         sfm.TotalEarned,
		TotalSpent:          sfm.TotalSpent,
		ReinvestmentHistory: sfm.ReinvestmentHistory,
	}
}

func (sfm *StackingFundManager) PaySubscription(service string, amount float64) {
	sfm.mu.Lock()
	defer sfm.mu.Unlock()

	sfm.Balance -= amount
	sfm.TotalSpent += amount

	sfm.Transactions = append(sfm.Transactions, Transaction{
		Timestamp:   time.Now(),
		Amount:      amount,
		Source:      service,
		Type:        "expense",
		Description: fmt.Sprintf("Subscription payment for %s", service),
	})
	sfm.save()
}

func (sfm *StackingFundManager) GetHistory(days int) []Transaction {
	sfm.mu.RLock()
	defer sfm.mu.RUnlock()

	var history []Transaction
	cutoff := time.Now().AddDate(0, 0, -days)
	for _, tx := range sfm.Transactions {
		if tx.Timestamp.After(cutoff) {
			history = append(history, tx)
		}
	}
	return history
}

func (sfm *StackingFundManager) ReinvestSurplus(threshold, percentage float64) {
	// ReinvestSurplus is handled by handleEconomics in RecordRevenue for the Stacking ecosystem
}

func (sfm *StackingFundManager) CoverStripeFees(transactionAmount float64) {
	sfm.mu.Lock()
	defer sfm.mu.Unlock()

	fee := (transactionAmount * 0.029) + 0.30
	sfm.Balance -= fee
	sfm.TotalSpent += fee

	sfm.Transactions = append(sfm.Transactions, Transaction{
		Timestamp:   time.Now(),
		Amount:      fee,
		Source:      "stripe",
		Type:        "expense",
		Description: fmt.Sprintf("Stripe fee for %.2f", transactionAmount),
	})
	sfm.save()
}
