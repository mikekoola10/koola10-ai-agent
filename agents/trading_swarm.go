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
	defer func() { a.status = StatusCompleted }()

	if a.specialty == "Arbitrage Scanner (DEX)" || a.specialty == "Arbitrage Scanner (CEX)" {
		// Use market data tool for arbitrage scanning
		res := tools.RunTool("market_data", map[string]interface{}{
			"symbol": "BTC/USD",
		})
		return res, nil
	}

	// Use existing crypto tool for paper trading
	res := tools.RunTool("crypto", map[string]interface{}{
		"action": "price",
		"symbol": "BTC",
	})

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
