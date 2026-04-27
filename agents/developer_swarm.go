package agents

import (
	"fmt"
)

type DeveloperAgent struct {
	specialty string
	status    AgentStatus
}

func (a *DeveloperAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	// Simulate development in isolated sandbox
	res := fmt.Sprintf("Completed %s task: %s", a.specialty, task)
	a.status = StatusCompleted
	return res, nil
}

func (a *DeveloperAgent) Status() AgentStatus { return a.status }
func (a *DeveloperAgent) Specialty() string    { return a.specialty }

func DeveloperFactory() []SpecialistAgent {
	specialties := []string{
		"Frontend (React)", "Frontend (Vue)", "Frontend (Svelte)",
		"Backend (Go)", "Backend (Python)", "Backend (Node)",
		"DevOps (Fly.io)", "DevOps (Docker)",
		"Testing Suite", "Documentation Generator",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &DeveloperAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
