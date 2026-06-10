package solara

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"koola10/financial"
	"koola10/vault"
)

type DailyBrief struct {
	ledger    *financial.FundManager
	vault     *vault.VaultClient
	apiKey    string
	notifier  func(string) error
}

func NewDailyBrief(ledger *financial.FundManager, vaultClient *vault.VaultClient, apiKey string, notify func(string) error) *DailyBrief {
	return &DailyBrief{
		ledger:    ledger,
		vault:     vaultClient,
		apiKey:    apiKey,
		notifier:  notify,
	}
}

func (db *DailyBrief) StartScheduler(ctx context.Context) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		log.Printf("[Solara] Next morning brief scheduled for %v", next)

		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			db.generateAndSend()
		}
	}
}

func (db *DailyBrief) generateAndSend() {
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

	brief, err := db.callDeepSeek(prompt)
	if err != nil {
		log.Printf("[Solara] Morning brief error: %v", err)
		return
	}

	if db.notifier != nil {
		if err := db.notifier(brief); err != nil {
			log.Printf("[Solara] Failed to send brief: %v", err)
		}
	}
	log.Printf("[Solara] Morning brief sent:\n%s", brief)
}

func (db *DailyBrief) callDeepSeek(prompt string) (string, error) {
	if db.apiKey == "" {
		return "", fmt.Errorf("DEEPSEEK_API_KEY not set")
	}

	dsReq := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	body, _ := json.Marshal(dsReq)
	req, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+db.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API error: %d", resp.StatusCode)
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
		return "", fmt.Errorf("no response from DeepSeek")
	}

	return dsRes.Choices[0].Message.Content, nil
}

func formatRecent(txs []financial.Transaction) string {
	if len(txs) == 0 {
		return "No transactions in last 24h."
	}
	lines := make([]string, len(txs))
	for i, tx := range txs {
		lines[i] = fmt.Sprintf("- %s: $%.2f (%s)", tx.Description, tx.Amount, tx.Type)
	}
	return strings.Join(lines, "\n")
}
