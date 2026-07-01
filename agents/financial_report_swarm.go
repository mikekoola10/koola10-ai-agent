package agents

import (
	"koola10/mirror"
)

type FinancialAgent struct {
	specialty string
	status    AgentStatus
	mirror    *mirror.Mirror
}

func (a *FinancialAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusIdle }()

	if a.mirror != nil {
		ctx := a.mirror.GetContext("sterling")
		_ = ctx.RiskTolerance
	}

	res := "Financial Report (" + a.specialty + "): " + task

	if a.mirror != nil {
		a.mirror.RecordOutcome("sterling", map[string]interface{}{"task": task, "success": true})
	}

	a.status = StatusCompleted
	return res, nil
}

func (a *FinancialAgent) Status() AgentStatus { return a.status }
func (a *FinancialAgent) Specialty() string    { return a.specialty }

func FinancialFactory(m *mirror.Mirror) func() []SpecialistAgent {
	return func() []SpecialistAgent {
	specialties := []string{
		"P&L Reporting (Monthly)", "P&L Reporting (Quarterly)",
		"Cash Flow Statement", "Cash Flow Forecasting",
		"Investor Update", "Investor Presentation",
		"Board Deck (Operational)", "Board Deck (Strategic)",
		"Variance Analysis (Budget)", "Variance Analysis (Actual)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
		for _, s := range specialties {
			agents = append(agents, &FinancialAgent{specialty: s, status: StatusIdle, mirror: m})
		}
		return agents
	}
}
