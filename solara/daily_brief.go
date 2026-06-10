package solara

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"koola10/financial"
	"koola10/vault"
)

type DailyBrief struct {
	ledger    financial.Ledger
	vault     *vault.VaultClient
	apiKey    string
	briefPath string
}

func NewDailyBrief(ledger financial.Ledger, vaultClient *vault.VaultClient, apiKey string) *DailyBrief {
	briefPath := "data/last_brief.txt"
	if _, err := os.Stat("/data"); err == nil {
		briefPath = "/data/last_brief.txt"
	}
	return &DailyBrief{
		ledger:    ledger,
		vault:     vaultClient,
		apiKey:    apiKey,
		briefPath: briefPath,
	}
}

func (db *DailyBrief) StartScheduler(ctx context.Context) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(next)):
			db.generateAndStore()
		}
	}
}

func (db *DailyBrief) generateAndStore() {
	total, ops, spend := db.ledger.GetRevenueSplit()
	recent := db.ledger.GetRecentTransactions(24 * time.Hour)
	prompt := fmt.Sprintf(`
Koola-10 Morning Brief – %s

Total revenue: $%.2f
Operations fund (bills): $%.2f
Spendable fund: $%.2f

Recent transactions (last 24h):
%s

Write a short, actionable briefing (3–4 sentences) highlighting changes, unpaid bills, and today’s top priority.
`, time.Now().Format("2006-01-02"), total, ops, spend, formatRecent(recent))

	brief, err := db.queryDeepSeek(prompt)
	if err != nil {
		log.Printf("[Solara] Morning brief error: %v", err)
		return
	}

	// Store in vault as type "brief"
	err = db.vault.AddEntry(vault.VaultEntry{
		Description: "Morning Brief " + time.Now().Format("2006-01-02"),
		Amount:      0,
		Type:        "brief",
		Notes:       brief,
	})
	if err != nil {
		log.Printf("[Solara] Failed to store brief in vault: %v", err)
	}

	// Also save to file for easy retrieval
	_ = os.MkdirAll(filepath.Dir(db.briefPath), 0755)
	os.WriteFile(db.briefPath, []byte(brief), 0644)
	log.Printf("[Solara] Morning brief stored:\n%s", brief)
}

func (db *DailyBrief) queryDeepSeek(prompt string) (string, error) {
	if db.apiKey == "" {
		return "DeepSeek API key not configured.", nil
	}

	reqBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+db.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api error: %s", resp.Status)
	}

	var dsRes struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsRes); err != nil {
		return "", err
	}

	if len(dsRes.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return dsRes.Choices[0].Message.Content, nil
}

func formatRecent(txs []financial.Transaction) string {
	if len(txs) == 0 {
		return "No transactions in last 24h."
	}
	lines := make([]string, len(txs))
	for i, tx := range txs {
		lines[i] = fmt.Sprintf("- %s: $%.2f", tx.Description, tx.Amount)
	}
	return strings.Join(lines, "\n")
}

// GetLatestBrief reads the brief from file (used by API handler)
func (db *DailyBrief) GetLatestBrief() (string, error) {
	data, err := os.ReadFile(db.briefPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
