package agents

import (
	"fmt"
	"log"
)

type MaintenanceAgent struct {
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *MaintenanceAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()
	log.Printf("[MaintenanceAgent] Specialty %s processing task: %s", a.specialty, task)
	return fmt.Sprintf("Maintenance task completed: %s", task), nil
}

func (a *MaintenanceAgent) Status() AgentStatus { return a.status }
func (a *MaintenanceAgent) Specialty() string    { return a.specialty }
func (a *MaintenanceAgent) SetPrompt(p string)   { a.prompt = p }
func (a *MaintenanceAgent) GetPrompt() string    { return a.prompt }

func MaintenanceFactory() []SpecialistAgent {
	specialties := []string{"diagnostics", "fix", "audit"}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &MaintenanceAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
