package agents

import (
	"fmt"
	"log"
)

type CopilotAgent struct {
	prompt    string
	specialty string
	status    AgentStatus
}

func (a *CopilotAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	log.Printf("[CopilotHive] Agent (%s) received task: %s", a.specialty, task)

	return fmt.Sprintf("Copilot Agent (%s) successfully processed task: %s", a.specialty, task), nil
}

func (a *CopilotAgent) Status() AgentStatus { return a.status }
func (a *CopilotAgent) Specialty() string    { return a.specialty }

func CopilotHiveFactory() []SpecialistAgent {
	specialties := []string{
		"Research (1)", "Research (2)", "Research (3)",
		"Research (4)", "Research (5)", "Research (6)",
		"Developer", "Auditor", "Emergency Fixer",
		"DevOps", "Security", "Performance", "Orchestrator",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &CopilotAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
func (a *CopilotAgent) SetPrompt(p string) { a.prompt = p }
