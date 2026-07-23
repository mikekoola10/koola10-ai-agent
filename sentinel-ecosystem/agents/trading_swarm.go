package agents

import (
	 "sentinel/tools"
)

type TradingAgent struct {
	specialty string
	status    AgentStatus
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

func TradingFactory() []SpecialistAgent {
	specialties := []string{
		"Arbitrage Scanner (DEX)", "Arbitrage Scanner (CEX)",
		"Grid Trading (Conservative)", "Grid Trading (Balanced)",
		"Risk Monitoring", "Stablecoin Yield",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &TradingAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
