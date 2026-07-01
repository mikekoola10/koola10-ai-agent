package agents

import (
	"koola10/mirror"
)

type GrantSwarmAgent struct {
	specialty string
	status    AgentStatus
	mirror    *mirror.Mirror
}

func (a *GrantSwarmAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusIdle }()

	if a.mirror != nil {
		ctx := a.mirror.GetContext("nova")
		_ = ctx.RiskTolerance
	}

	res := "Grant Proposal (" + a.specialty + "): " + task

	if a.mirror != nil {
		a.mirror.RecordOutcome("nova", map[string]interface{}{"task": task, "success": true})
	}

	a.status = StatusCompleted
	return res, nil
}

func (a *GrantSwarmAgent) Status() AgentStatus { return a.status }
func (a *GrantSwarmAgent) Specialty() string    { return a.specialty }

func GrantSwarmFactory(m *mirror.Mirror) func() []SpecialistAgent {
	return func() []SpecialistAgent {
	specialties := []string{
		"Federal Database Monitor", "Federal Proposal Draft", "Federal Compliance",
		"State Grant Search", "State Proposal Draft", "State Budget Plan",
		"Foundation Outreach", "Foundation Proposal", "Private Grant Search", "Impact Report Gen",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
		for _, s := range specialties {
			agents = append(agents, &GrantSwarmAgent{specialty: s, status: StatusIdle, mirror: m})
		}
		return agents
	}
}
