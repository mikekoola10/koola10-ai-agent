package agents

type GrantSwarmAgent struct {
	manager *SwarmManager
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *GrantSwarmAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Grant Proposal (" + a.specialty + "): " + a.prompt + " | Task: " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *GrantSwarmAgent) Status() AgentStatus { return a.status }
func (a *GrantSwarmAgent) Specialty() string    { return a.specialty }
func (a *GrantSwarmAgent) SetPrompt(p string)   { a.prompt = p }
func (a *GrantSwarmAgent) GetPrompt() string    { return a.prompt }

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

func (a *GrantSwarmAgent) SetManager(m *SwarmManager) { a.manager = m }

func (a *GrantSwarmAgent) ConfidenceLevel() float64 { return 0.95 }
func (a *GrantSwarmAgent) RequestClarification(ctx string) string { return "" }
