package agents

type BountyAgent struct {
	specialty string
	status    AgentStatus
}

func (a *BountyAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	// Simulate bug bounty hunting or task completion
	a.status = StatusCompleted
	return map[string]interface{}{
		"status": "success",
		"message": "Bounty task completed: " + a.specialty,
		"profit": 150.0, // Projected profit
	}, nil
}

func (a *BountyAgent) Status() AgentStatus { return a.status }
func (a *BountyAgent) Specialty() string    { return a.specialty }

func BountyFactory() []SpecialistAgent {
	specialties := []string{
		"Security Researcher", "Code Auditor", "Vulnerability Scanner",
		"Bug Hunter (Web)", "Bug Hunter (Mobile)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BountyAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
