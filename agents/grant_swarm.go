package agents

type GrantSwarmAgent struct {
	prompt    string
	specialty string
	status    AgentStatus
}

func (a *GrantSwarmAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Grant Proposal (" + a.specialty + "): " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *GrantSwarmAgent) Status() AgentStatus { return a.status }
func (a *GrantSwarmAgent) Specialty() string    { return a.specialty }

func GrantSwarmFactory() []SpecialistAgent {
	specialties := []string{
		"Federal Database Monitor", "Federal Proposal Draft", "Federal Compliance",
		"State Grant Search", "State Proposal Draft", "State Budget Plan",
		"Foundation Outreach", "Foundation Proposal", "Private Grant Search", "Impact Report Gen",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &GrantSwarmAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
func (a *GrantSwarmAgent) SetPrompt(p string) { a.prompt = p }
