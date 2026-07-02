package agents

import (
	"fmt"
	"math/rand"
	"time"
)

type AffiliateAgent struct {
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *AffiliateAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// Simulate deep research and article generation
	time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

	// Simulate finding a profit (commission)
	profit := 10.0 + rand.Float64()*40.0 // $10 - $50

	res := map[string]interface{}{
		"article": fmt.Sprintf("Deep research article for: %s", task),
		"profit":  profit,
		"status":  "published",
	}

	return res, nil
}

func (a *AffiliateAgent) Status() AgentStatus { return a.status }
func (a *AffiliateAgent) Specialty() string    { return a.specialty }
func (a *AffiliateAgent) SetPrompt(p string)   { a.prompt = p }
func (a *AffiliateAgent) GetPrompt() string    { return a.prompt }

func AffiliateFactory() []SpecialistAgent {
	specialties := []string{
		"Tech Reviewer", "Finance Blogger", "Health Guru", "Travel Expert", "AI Enthusiast",
		"Gadget Specialist", "Lifestyle Curator", "Business Analyst", "Gaming Critic", "Home Decorator",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &AffiliateAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
