package agents

type BountyAgent struct {
	specialty string
	status    AgentStatus
}

func (a *BountyAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Bounty Result (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *BountyAgent) Status() AgentStatus { return a.status }
func (a *BountyAgent) Specialty() string    { return a.specialty }

func BountyFactory() []SpecialistAgent {
	specialties := []string{
		"Bug Bounty Scanner", "HackerOne Lead Scout", "Bugcrowd Program Monitor",
		"Vulnerability Reporter", "Exploit Researcher", "Payload Generator",
		"Automated Fuzzer", "Code Auditor", "Bounty Payment Tracker", "Responsible Disclosure Manager",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BountyAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
