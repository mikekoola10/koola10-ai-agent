package agents

type FinancialAgent struct {
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *FinancialAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	res := "Financial Report (" + a.specialty + "): " + a.prompt + " | Task: " + task
	a.status = StatusCompleted
	return res, nil
}

func (a *FinancialAgent) Status() AgentStatus { return a.status }
func (a *FinancialAgent) Specialty() string    { return a.specialty }
func (a *FinancialAgent) SetPrompt(p string)   { a.prompt = p }
func (a *FinancialAgent) GetPrompt() string    { return a.prompt }

func FinancialFactory() []SpecialistAgent {
	specialties := []string{
		"P&L Reporting (Monthly)", "P&L Reporting (Quarterly)",
		"Cash Flow Statement", "Cash Flow Forecasting",
		"Investor Update", "Investor Presentation",
		"Board Deck (Operational)", "Board Deck (Strategic)",
		"Variance Analysis (Budget)", "Variance Analysis (Actual)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &FinancialAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
