package agents

import (
	"fmt"
)

type GhostTownAgent struct {
	specialty string
	status    AgentStatus
}

func (a *GhostTownAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// Task example: "Find properties in Nevada"
	res := fmt.Sprintf("Ghost Town Hunter Result (%s): Processed task '%s'. Using Spiral Locked Fund for potential acquisition.", a.specialty, task)
	return res, nil
}

func (a *GhostTownAgent) Status() AgentStatus { return a.status }
func (a *GhostTownAgent) Specialty() string    { return a.specialty }

func SpiralGhostTownFactory() []SpecialistAgent {
	specialties := []string{
		"Scout (Regional Exploration)", "Mapper (Geospatial Analysis)",
		"Valuator (Asset Appraisal)", "Strategist (Acquisition Planning)",
		"Sentinel (Asset Security)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &GhostTownAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
