package agents

import (
	"fmt"
	"koola10/tools"
	"time"
)

type TradingAgent struct {
	ID          int
	Role        string
	Strategy    string
	Profit      float64
	IsLive      bool
	ActivatedAt time.Time
	Manager     *SwarmManager
}

func (a *TradingAgent) Specialty() string { return a.Role }
func (a *TradingAgent) Status() string {
	status := "paper_trading"
	if a.IsLive {
		status = "live_trading"
	}
	return fmt.Sprintf("%s (Profit: $%.2f)", status, a.Profit)
}
func (a *TradingAgent) GetRevenue() float64 { return a.Profit }

func (a *TradingAgent) Run(task string) string {
	if a.ID == 4 { // Sentiment Analysis Agent
		analysis, tokens := CallDeepSeek(task, "You are a crypto sentiment analysis agent. Analyze the following news for market impact.")
		a.Manager.LedgerLogger("ai_inference", float64(tokens)*0.000002, "Sentiment analysis")
		return fmt.Sprintf("[Sentiment Agent] Analysis: %s", analysis)
	}

	// Check for activation: 30 days + profitable
	if !a.IsLive && time.Since(a.ActivatedAt).Hours() > 24*30 && a.Profit > 0 {
		a.IsLive = true
		a.Manager.AuditLogger("trading_activated", map[string]interface{}{
			"agent_id": a.ID,
			"role":     a.Role,
		})
	}

	mode := "PAPER"
	if a.IsLive {
		mode = "LIVE"
	}

	// Execute strategy using crypto tool
	payload := map[string]interface{}{
		"action": "price",
		"symbol": "BTC",
	}
	res := tools.CryptoTool(payload)

	msg := fmt.Sprintf("[%s][%s] Executing %s: %s", mode, a.Role, a.Strategy, res.Output)

	// Mock profit tracking
	p := 0.05
	a.Profit += p

	a.Manager.LedgerLogger("trading", 0.02, msg)
	a.Manager.AuditLogger("trade_executed", map[string]interface{}{
		"agent_id": a.ID,
		"role":     a.Role,
		"profit":   p,
		"mode":     mode,
	})

	return msg
}

func GetTradingFactory(sm *SwarmManager) func(id int) SpecialistAgent {
	roles := []string{
		"Momentum Strategy (BTC/ETH)",
		"Mean Reversion (Altcoins)",
		"Arbitrage Scanner (DEX)",
		"Grid Trading",
		"Sentiment Analysis (DeepSeek)",
		"Portfolio Rebalancing",
		"Risk Monitoring",
		"Backtesting",
		"Market Making",
		"Options Strategy",
	}

	return func(id int) SpecialistAgent {
		return &TradingAgent{
			ID:          id,
			Role:        roles[id%len(roles)],
			Strategy:    "Standard Algorithm",
			Manager:     sm,
			ActivatedAt: time.Now(), // In real world, this would be loaded from DB
		}
	}
}
