package agents

import (
	"fmt"
	"log"
	"strings"
)

type MaintenanceAgent struct {
	prompt    string
	specialty string
	status    AgentStatus
}

func (a *MaintenanceAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	log.Printf("[MaintenanceAgent] Specialty %s processing task: %s", a.specialty, task)

	switch a.specialty {
	case "diagnostics":
		return fmt.Sprintf("Diagnostics complete for: %s. Identified potential root cause: dependency timeout.", task), nil
	case "fix":
		return fmt.Sprintf("Self-healing fix applied for: %s. Restarting failed service.", task), nil
	case "audit":
		return fmt.Sprintf("System audit complete after fix for: %s. Status: RECOVERED.", task), nil
	default:
		if strings.Contains(task, "repair") {
			return fmt.Sprintf("General repair initiated for: %s", task), nil
		}
		return fmt.Sprintf("Maintenance task completed: %s", task), nil
	}
}

func (a *MaintenanceAgent) Status() AgentStatus { return a.status }
func (a *MaintenanceAgent) Specialty() string    { return a.specialty }

func MaintenanceFactory() []SpecialistAgent {
	specialties := []string{
		"diagnostics", "fix", "audit",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &MaintenanceAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
func (a *MaintenanceAgent) SetPrompt(p string) { a.prompt = p }
