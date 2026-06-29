package solara

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"koola10/financial"
)

type DailyBrief struct {
	ledger        *financial.EconomicLedger
	vaultClient   interface{}
	briefFilePath string
}

func NewDailyBrief(ledger *financial.EconomicLedger, briefPath string) *DailyBrief {
	return &DailyBrief{
		ledger:        ledger,
		briefFilePath: briefPath,
	}
}

func (db *DailyBrief) StartScheduler() {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}

			log.Printf("[Solara] Next morning brief scheduled for: %v", next)
			time.Sleep(time.Until(next))

			db.GenerateAndStore()
		}
	}()
}

func (db *DailyBrief) GenerateAndStore() {
	log.Printf("[Solara] Generating daily morning brief...")

	total, ops, spend := db.ledger.GetRevenueSplit()
	recent := db.ledger.GetRecentTransactions(24)

	summary := fmt.Sprintf("Total Revenue: %.2f. Operations Fund: %.2f. Spendable Fund: %.2f. Recent Transactions: %d.", total, ops, spend, len(recent))

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	brief := "Good morning. " + summary // Fallback

	if apiKey != "" {
		prompt := fmt.Sprintf("Write a short morning briefing (3-4 sentences) for an AI agent swarm based on these stats: %s. Be concise and professional.", summary)

		dsReq := map[string]interface{}{
			"model": "deepseek-chat",
			"messages": []map[string]string{
				{"role": "system", "content": "You are Solara, the content agent for Koola10."},
				{"role": "user", "content": prompt},
			},
		}

		body, _ := json.Marshal(dsReq)
		req, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			var dsRes struct {
				Choices []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&dsRes); err == nil && len(dsRes.Choices) > 0 {
				brief = dsRes.Choices[0].Message.Content
			}
		} else {
			log.Printf("[Solara] DeepSeek API failed: %v", err)
		}
	}

	// Store in file
	os.WriteFile(db.briefFilePath, []byte(brief), 0644)

	// Store in vault (using ledger as placeholder for vault record)
	db.ledger.RecordTransaction("Morning Brief", 0, "brief", brief)

	log.Printf("[Solara] Morning brief generated and stored.")
}

func (db *DailyBrief) GetLatestBrief() string {
	data, err := os.ReadFile(db.briefFilePath)
	if err != nil {
		return "No brief available yet. Check back at 8 AM."
	}
	return string(data)
}
