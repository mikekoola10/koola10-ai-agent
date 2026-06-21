package agents

import (
	"fmt"
	"koola10/mirror"
	"koola10/tools"
)

type TradingAgent struct {
	specialty string
	status    AgentStatus
	mirror    *mirror.Mirror
}

func (a *TradingAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusIdle }()

	// APEX Proactive logic
	if a.mirror != nil {
		if val, ok := a.mirror.GetPreference("trading", "risk_appetite"); ok {
			fmt.Printf("[APEX] Trading Agent consulting mirror: Risk Appetite is %s\n", val)
		}
	}

	// Use existing crypto tool for paper trading
	res := tools.RunTool("crypto", map[string]interface{}{
		"action": "price",
		"symbol": "BTC",
	})

	a.status = StatusCompleted

	if a.mirror != nil {
		a.mirror.RecordOutcome("trading", map[string]interface{}{"task": task, "success": true})
	}

	return res, nil
}

func (a *TradingAgent) Status() AgentStatus { return a.status }
func (a *TradingAgent) Specialty() string    { return a.specialty }

func TradingFactory(m *mirror.Mirror) func() []SpecialistAgent {
	return func() []SpecialistAgent {
	specialties := []string{
		"Momentum Strategy (1m)", "Momentum Strategy (5m)", "Momentum Strategy (15m)",
		"Mean Reversion (Bollinger)", "Mean Reversion (RSI)",
		"Arbitrage Scanner (DEX)", "Arbitrage Scanner (CEX)",
		"Sentiment Analysis", "Portfolio Rebalancing", "Risk Monitoring",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
		for _, s := range specialties {
			agents = append(agents, &TradingAgent{specialty: s, status: StatusIdle, mirror: m})
		}
		return agents
	}
}
