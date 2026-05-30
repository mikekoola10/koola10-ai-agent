package agents

type ResearchAgent struct {
	specialty string
	status    AgentStatus
}

func (a *ResearchAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Market Intel (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *ResearchAgent) Status() AgentStatus { return a.status }
func (a *ResearchAgent) Specialty() string    { return a.specialty }

func ResearchFactory() []SpecialistAgent {
	specialties := []string{
		"Risk Intelligence (Cyber)", "Risk Intelligence (Financial)",
		"Threat Monitoring (Market)", "Threat Monitoring (Operational)",
		"Compliance Trend Analysis", "Intelligence Brief Generator",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &ResearchAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
