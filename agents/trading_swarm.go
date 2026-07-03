package agents

import (
	"koola10/tools"
)

type TradingAgent struct {
	manager *SwarmManager
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *TradingAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	// Use existing crypto tool for paper trading
	res := tools.RunTool("crypto", map[string]interface{}{
		"action": "price",
		"symbol": "BTC",
	})

	a.status = StatusCompleted
	return res, nil
}

func (a *TradingAgent) Status() AgentStatus { return a.status }
func (a *TradingAgent) Specialty() string    { return a.specialty }


func (a *TradingAgent) SetPrompt(p string)   { a.prompt = p }
func (a *TradingAgent) GetPrompt() string    { return a.prompt }

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

func (a *TradingAgent) SetManager(m *SwarmManager) { a.manager = m }
