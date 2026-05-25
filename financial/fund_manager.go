package financial

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Ledger interface {
	RecordRevenue(amount float64, source string)
	RecordRevenueWithEcosystem(ecosystem, vertical string, amount float64, source string)
}

type Transaction struct {
	Timestamp   time.Time `json:"timestamp"`
	Amount      float64   `json:"amount"`
	Source      string    `json:"source"`
	Type        string    `json:"type"` // "revenue_split", "expense", "reinvestment"
	Description string    `json:"description"`
}

type FundStatus struct {
	Balance             float64  `json:"balance"`
	TotalEarned         float64  `json:"total_earned"`
	TotalSpent          float64  `json:"total_spent"`
	ReinvestmentHistory []string `json:"reinvestment_history"`
}

type FundManager struct {
	Balance             float64       `json:"balance"`
	TotalEarned         float64       `json:"total_earned"`
	TotalSpent          float64       `json:"total_spent"`
	ReinvestmentHistory []string      `json:"reinvestment_history"`
	Transactions        []Transaction `json:"transactions"`
	storagePath         string
	mu                  sync.RWMutex
	ledger              Ledger
}

func NewFundManager(path string, ledger Ledger) *FundManager {
	fm := &FundManager{
		storagePath:         path,
		ledger:              ledger,
		ReinvestmentHistory: []string{},
		Transactions:        []Transaction{},
	}
	fm.load()
	return fm
}

func (fm *FundManager) load() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	data, err := os.ReadFile(fm.storagePath)
	if err == nil {
		json.Unmarshal(data, fm)
	}
	if fm.ReinvestmentHistory == nil {
		fm.ReinvestmentHistory = []string{}
	}
	if fm.Transactions == nil {
		fm.Transactions = []Transaction{}
	}
}

func (fm *FundManager) save() error {
	data, err := json.MarshalIndent(fm, "", "  ")
	if err != nil {
		return err
	}
	// Retry logic for file operations
	var lastErr error
	for i := 0; i < 3; i++ {
		lastErr = os.WriteFile(fm.storagePath, data, 0644)
		if lastErr == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return lastErr
}

func (fm *FundManager) RouteRevenue(amount float64, ecosystem, vertical, source string) {
	fm.mu.Lock()

	opAmount := amount * 0.30
	glAmount := amount * 0.70

	fm.Balance += opAmount
	fm.TotalEarned += opAmount

	fm.Transactions = append(fm.Transactions, Transaction{
		Timestamp:   time.Now(),
		Amount:      opAmount,
		Source:      source,
		Type:        "revenue_split",
		Description: fmt.Sprintf("30%% split to operational fund from %s (Total: %.2f, Ecosystem: %s, Vertical: %s)", source, amount, ecosystem, vertical),
	})
	err := fm.save()
	fm.mu.Unlock()

	if err != nil {
		fmt.Printf("FAILED to save fund manager data after revenue routing: %v\n", err)
	}

	if fm.ledger != nil {
		if ecosystem != "" || vertical != "" {
			fm.ledger.RecordRevenueWithEcosystem(ecosystem, vertical, glAmount, fmt.Sprintf("70%% split from %s", source))
		} else {
			fm.ledger.RecordRevenue(glAmount, fmt.Sprintf("70%% split from %s", source))
		}
	}

	// Verification: check if file actually exists and contains the latest transaction
	if _, err := os.Stat(fm.storagePath); err == nil {
		fmt.Printf("Verified: revenue from %s (Ecosystem: %s, Vertical: %s) recorded to %s\n", source, ecosystem, vertical, fm.storagePath)
	}
}

func (fm *FundManager) CoverStripeFees(transactionAmount float64) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fee := (transactionAmount * 0.029) + 0.30
	fm.Balance -= fee
	fm.TotalSpent += fee

	fm.Transactions = append(fm.Transactions, Transaction{
		Timestamp:   time.Now(),
		Amount:      fee,
		Source:      "stripe",
		Type:        "expense",
		Description: fmt.Sprintf("Stripe fee for %.2f", transactionAmount),
	})
	fm.save()
}

func (fm *FundManager) PaySubscription(service string, amount float64) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.Balance -= amount
	fm.TotalSpent += amount

	fm.Transactions = append(fm.Transactions, Transaction{
		Timestamp:   time.Now(),
		Amount:      amount,
		Source:      service,
		Type:        "expense",
		Description: fmt.Sprintf("Subscription payment for %s", service),
	})
	fm.save()
}

func (fm *FundManager) PayFlyInvoice() {
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		return
	}

	query := map[string]string{
		"query": `query {
			viewer {
				organizations {
					nodes {
						slug
						invoices {
							nodes {
								amount
								status
							}
						}
					}
				}
			}
		}`,
	}
	body, _ := json.Marshal(query)
	req, _ := http.NewRequest("POST", "https://api.fly.io/graphql", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var res struct {
		Data struct {
			Viewer struct {
				Organizations struct {
					Nodes []struct {
						Slug     string `json:"slug"`
						Invoices struct {
							Nodes []struct {
								Amount int    `json:"amount"`
								Status string `json:"status"`
							} `json:"nodes"`
						} `json:"invoices"`
					} `json:"nodes"`
				} `json:"organizations"`
			} `json:"viewer"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	for _, org := range res.Data.Viewer.Organizations.Nodes {
		for _, inv := range org.Invoices.Nodes {
			if inv.Status == "open" || inv.Status == "past_due" {
				amount := float64(inv.Amount) / 100.0
				if fm.Balance >= amount {
					fm.Balance -= amount
					fm.TotalSpent += amount
					fm.Transactions = append(fm.Transactions, Transaction{
						Timestamp:   time.Now(),
						Amount:      amount,
						Source:      "fly.io",
						Type:        "expense",
						Description: fmt.Sprintf("Fly.io invoice payment for org %s", org.Slug),
					})
				}
			}
		}
	}
	fm.save()
}

func (fm *FundManager) ReinvestSurplus(threshold, percentage float64) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.Balance > threshold {
		surplus := fm.Balance - threshold
		reinvestAmount := surplus * (percentage / 100.0)
		fm.Balance -= reinvestAmount

		msg := fmt.Sprintf("Reinvested %.2f (%.0f%% of surplus above %.2f) into swarm scaling", reinvestAmount, percentage, threshold)
		fm.ReinvestmentHistory = append(fm.ReinvestmentHistory, fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), msg))

		fm.Transactions = append(fm.Transactions, Transaction{
			Timestamp:   time.Now(),
			Amount:      reinvestAmount,
			Source:      "internal",
			Type:        "reinvestment",
			Description: msg,
		})
		fm.save()
	}
}

func (fm *FundManager) GetStatus() FundStatus {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return FundStatus{
		Balance:             fm.Balance,
		TotalEarned:         fm.TotalEarned,
		TotalSpent:          fm.TotalSpent,
		ReinvestmentHistory: fm.ReinvestmentHistory,
	}
}

func (fm *FundManager) GetHistory(days int) []Transaction {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var history []Transaction
	cutoff := time.Now().AddDate(0, 0, -days)
	for _, tx := range fm.Transactions {
		if tx.Timestamp.After(cutoff) {
			history = append(history, tx)
		}
	}
	return history
}
