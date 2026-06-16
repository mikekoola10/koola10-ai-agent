package agents

import (
	"fmt"
	"log"
)

type BuilderAgent struct {
	specialty string
	status    AgentStatus
}

func (a *BuilderAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	log.Printf("[Builder] %s is constructing artifact: %s", a.specialty, task)
	return fmt.Sprintf("Builder (%s) produced artifact for: %s", a.specialty, task), nil
}

func (a *BuilderAgent) Status() AgentStatus { return a.status }
func (a *BuilderAgent) Specialty() string    { return a.specialty }

func BuilderFactory() []SpecialistAgent {
	specialties := []string{
		"App Scaffolder", "CI/CD Pipeline Builder", "Microservice Weaver",
		"Database Schema Designer", "Integration Architect",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BuilderAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
