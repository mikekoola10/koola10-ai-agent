package agents

import (
	"math/rand"
	"time"
)

type BountyAgent struct {
	specialty string
	status    AgentStatus
	prompt    string
	manager   *SwarmManager
}

func (a *BountyAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// Simulate scanning target with AGI context
	time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)

	// Simulate finding a bug and a bounty
	foundProbability := 0.3
	if a.manager != nil && a.manager.IsAGIMode() {
		foundProbability = 0.5 // AGI mode increases discovery rate
	}

	found := rand.Float64() < foundProbability
	profit := 0.0
	if found {
		baseProfit := 100.0 + rand.Float64()*900.0
		if foundProbability > 0.3 {
			baseProfit *= 2.0 // 10x thinking leads to higher impact bugs
		}
		profit = baseProfit
	}

	res := map[string]interface{}{
		"target":      task,
		"vulnerable":  found,
		"profit":      profit,
		"report_sent": found,
		"intelligence": "AGI-enhanced scan complete",
	}

	return res, nil
}

func (a *BountyAgent) Status() AgentStatus { return a.status }
func (a *BountyAgent) Specialty() string    { return a.specialty }
func (a *BountyAgent) SetPrompt(prompt string) { a.prompt = prompt }
func (a *BountyAgent) GetPrompt() string    { return "bounty" }

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

func (a *BountyAgent) SetManager(m *SwarmManager) {
	a.manager = m
}
