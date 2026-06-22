package agents

import (
	"fmt"
	"math/rand"
)

type BountyAgent struct {
	specialty string
	status    AgentStatus
	vertical  string
}

func (a *BountyAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// Simulate bug bounty hunting or task completion
	profit := 50.0 + rand.Float64()*500.0

	res := fmt.Sprintf("Bounty Result (%s - %s): Secured %.2f bounty from task: %s", a.vertical, a.specialty, profit, task)
	return res, nil
}

func (a *BountyAgent) Status() AgentStatus { return a.status }
func (a *BountyAgent) Specialty() string    { return a.specialty }

func BountyFactory() []SpecialistAgent {
	return CreateBountySwarm("bounty")
}

func SpiralBountyFactory() []SpecialistAgent {
	return CreateBountySwarm("spiral_bounty")
}

func CreateBountySwarm(vertical string) []SpecialistAgent {
	specialties := []string{
		"HackerOne Scanner", "Bugcrowd Vulnerability Finder", "Gitcoin Task Automator",
		"Subdomain Brute-forcer", "XSS Payload Tester", "SQLi Auditor",
		"Documentation Auditor (Bounty)", "Code Reviewer (Security)",
		"Responsible Disclosure Manager", "Payload Optimizer",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &BountyAgent{specialty: s, status: StatusIdle, vertical: vertical})
	}
	return agents
}
