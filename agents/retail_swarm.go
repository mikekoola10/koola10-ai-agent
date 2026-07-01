package agents

type RetailAgent struct {
	specialty string
	status    AgentStatus
}

func (a *RetailAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Retail Op (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *RetailAgent) Status() AgentStatus { return a.status }
func (a *RetailAgent) Specialty() string    { return a.specialty }

func RetailSwarmFactory() []SpecialistAgent {
	specialties := []string{
		"Inventory Management (Nova)",
		"Dynamic Pricing (Sterling)",
		"Staff Scheduling (Forge)",
		"Digital Signage (Solara)",
		"Loss Prevention (Sage)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &RetailAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
