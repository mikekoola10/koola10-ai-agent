package agents

import (
	"koola10/tools"
)

type TradingAgent struct {
	specialty string
	status    AgentStatus
}

func (a *TradingAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	// Use the new binance tool for real market data in paper trading mode
	res := tools.RunTool("binance", map[string]interface{}{
		"action": "get_price",
		"symbol": "BTCUSDT",
	})

	a.status = StatusCompleted
	return res, nil
}

func (a *TradingAgent) Status() AgentStatus { return a.status }
func (a *TradingAgent) Specialty() string    { return a.specialty }

func TradingFactory() []SpecialistAgent {
	specialties := []string{
		"Momentum Strategy (1m)", "Momentum Strategy (5m)", "Momentum Strategy (15m)",
		"Mean Reversion (Bollinger)", "Mean Reversion (RSI)",
		"Arbitrage Scanner (DEX)", "Arbitrage Scanner (CEX)",
		"Sentiment Analysis", "Portfolio Rebalancing", "Risk Monitoring",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &TradingAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
