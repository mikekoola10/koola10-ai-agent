package agents

type GenericArbitrageAgent struct {
	specialty string
	status    AgentStatus
}

func (a *GenericArbitrageAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	// Opportunity scouting logic
	a.status = StatusCompleted
	return "arbitrage opportunities found", nil
}

func (a *GenericArbitrageAgent) Status() AgentStatus { return a.status }
func (a *GenericArbitrageAgent) Specialty() string    { return a.specialty }

func ArbitrageFactory() []SpecialistAgent {
	specialties := []string{"Market Scout", "ROI Evaluator", "Bid Closer"}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &GenericArbitrageAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
