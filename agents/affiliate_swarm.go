package agents

import (
	"fmt"
	"math/rand"
	"time"
)

type AffiliateAgent struct {
	specialty string
	status    AgentStatus
}

func (a *AffiliateAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	time.Sleep(2 * time.Second) // Simulate article generation

	revenue := 0.0
	if rand.Float64() < 0.1 { // 10% chance of immediate click/sale simulation
		revenue = 10.0 + rand.Float64()*50.0
	}

	a.status = StatusCompleted
	return map[string]interface{}{
		"article_title": fmt.Sprintf("Top AI Tools for %s in 2026", task),
		"platform":      "Medium",
		"simulated_revenue": revenue,
	}, nil
}

func (a *AffiliateAgent) Status() AgentStatus { return a.status }
func (a *AffiliateAgent) Specialty() string    { return a.specialty }

func AffiliateFactory() []SpecialistAgent {
	specialties := []string{
		"AI Tool Reviewer", "SaaS Comparison Expert", "Tech Trend Analyst",
		"SEO Optimizer", "Affiliate Link Manager",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &AffiliateAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
