package agents

import (
	"fmt"
	"log"
)

type GBaseAgent struct {
	prompt    string
	specialty string
	status    AgentStatus
}

func (a *GBaseAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	log.Printf("[GBase] Agent (%s) analyzing task: %s", a.specialty, task)

	switch a.specialty {
	case "reflector":
		// Analyzes task outcomes and identifies failures/successes
		return fmt.Sprintf("Reflector analyzed: %s. Insight: Identified failure due to ambiguous prompt.", task), nil
	case "strategist":
		// Proposes prompt or strategy updates based on reflections
		return fmt.Sprintf("Strategist proposed update for: %s. Suggestion: Include explicit ROI constraints in system prompt.", task), nil
	default:
		return fmt.Sprintf("GBase task completed: %s", task), nil
	}
}

func (a *GBaseAgent) Status() AgentStatus { return a.status }
func (a *GBaseAgent) Specialty() string    { return a.specialty }

func GBaseFactory() []SpecialistAgent {
	specialties := []string{
		"reflector", "strategist",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &GBaseAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
func (a *GBaseAgent) SetPrompt(p string) { a.prompt = p }
