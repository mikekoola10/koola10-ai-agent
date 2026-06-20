package agents

import (
	"math/rand"
	"time"
)

type BountyAgent struct {
	specialty string
	status    AgentStatus
}

func (a *BountyAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	time.Sleep(3 * time.Second) // Simulate scanning

	found := false
	reward := 0.0
	if rand.Float64() < 0.05 { // 5% chance of finding a bug
		found = true
		reward = 100.0 + rand.Float64()*1000.0
	}

	a.status = StatusCompleted
	return map[string]interface{}{
		"target": task,
		"found_vulnerability": found,
		"estimated_reward": reward,
		"severity": "Medium",
	}, nil
}

func (a *BountyAgent) Status() AgentStatus { return a.status }
func (a *BountyAgent) Specialty() string    { return a.specialty }

func BountyFactory() []SpecialistAgent {
	specialties := []string{
		"XSS Scanner", "SQLi Auditor", "Auth Bypass Specialist",
		"Recon Specialist", "Report Writer",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BountyAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
