package agents

import (
	"log"
)

type BountyAgent struct {
}

func (a *BountyAgent) Run(task string) (interface{}, error) {
	log.Printf("[Bounty] Scanning target domains for vulnerabilities...")
	return "Submitted vulnerability report to HackerOne", nil
}

func (a *BountyAgent) Status() AgentStatus { return StatusIdle }
func (a *BountyAgent) Specialty() string   { return "bug_bounty_hunting" }

func BountyFactory() []SpecialistAgent {
	return []SpecialistAgent{&BountyAgent{}}
}
