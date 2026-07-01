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
	defer func() { a.status = StatusCompleted }()

	// Simulate scanning target
	time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)

	// Simulate finding a bug and a bounty
	found := rand.Float64() < 0.3 // 30% chance
	profit := 0.0
	if found {
		profit = 100.0 + rand.Float64()*900.0 // $100 - $1000
	}

	res := map[string]interface{}{
		"target":      task,
		"vulnerable":  found,
		"profit":      profit,
		"report_sent": found,
	}

	return res, nil
}

func (a *BountyAgent) Status() AgentStatus { return a.status }
func (a *BountyAgent) Specialty() string    { return a.specialty }

func BountyFactory() []SpecialistAgent {
	specialties := []string{
		"Web Pentester", "API Auditor", "Mobile Security", "Cloud Architect", "Network Scanner",
		"Auth Specialist", "Encryption Expert", "Logic Bug Hunter", "Zero-Day Researcher", "Report Writer",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BountyAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
