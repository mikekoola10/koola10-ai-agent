package agents

import (
	"log"
	"koola10/tools"
)

type RepurposeAgent struct {
	specialty string
	status    AgentStatus
}

func (a *RepurposeAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	log.Printf("[RepurposeAgent] Transforming content: %s", task)

	// 1. Summarize content via 9Router
	tools.RunTool("9router", map[string]interface{}{
		"action": "route",
		"prompt": "Repurpose this for Twitter/LinkedIn: " + task,
		"priority": "low",
	})

	return map[string]interface{}{
		"status": "success",
		"repurposed_content": "Simulated thread/post for: " + task,
	}, nil
}

func (a *RepurposeAgent) Status() AgentStatus { return a.status }
func (a *RepurposeAgent) Specialty() string    { return a.specialty }

func RepurposeFactory() []SpecialistAgent {
	specialties := []string{
		"Tweet Thread Generator", "LinkedIn Post Formatter",
		"Video Script Writer", "Infographic Outline Creator",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &RepurposeAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
