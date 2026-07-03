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
	manager   *SwarmManager
}

func (a *AffiliateAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// Incorporate AGI prompt if present for advanced reasoning simulation
	logMsg := fmt.Sprintf("Executing task: %s", task)
	agiActive := false
	if a.manager != nil && a.manager.IsAGIMode() {
		agiActive = true
		logMsg = fmt.Sprintf("Executing task with AGI optimization: %s", task)
	}

	if a.prompt != "" {
		logMsg = fmt.Sprintf("Executing task with directive: %s\nTask: %s", a.prompt, task)
	}

	// Simulate deep research and article generation
	time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

	// Simulate finding a profit (commission) with potential 10x leverage
	multiplier := 1.0
	if agiActive {
		multiplier = 1.5 // 50% boost from AGI optimization
	}
	profit := (10.0 + rand.Float64()*40.0) * multiplier

	res := map[string]interface{}{
		"article": fmt.Sprintf("%s\nResult: Deep research article generated with high agency.", logMsg),
		"profit":  profit,
		"status":  "published",
	}

	return res, nil
}

func (a *AffiliateAgent) Status() AgentStatus { return a.status }
func (a *AffiliateAgent) Specialty() string    { return a.specialty }
func (a *AffiliateAgent) SetPrompt(prompt string) { a.prompt = prompt }
func (a *AffiliateAgent) GetPrompt() string    { return "affiliate" }

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

// SetManager links the agent to the manager for state checks
func (a *AffiliateAgent) SetManager(m *SwarmManager) {
	a.manager = m
}
