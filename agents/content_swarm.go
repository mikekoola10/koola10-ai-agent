package agents

import (
	"fmt"
	"koola10/tools"
)

type ContentAgent struct {
	manager *SwarmManager
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *ContentAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	res := "Content Result (" + a.specialty + "): " + a.prompt + " | Task: " + task

	// AGI Mode: Integrate Hugging Face for Creative Breakthroughs
	if a.manager != nil && a.manager.IsAGIMode() {
		hfRes := tools.RunTool("huggingface", map[string]interface{}{
			"action": "run_model",
			"model":  "stable-diffusion-xl-base-1.0",
			"inputs": fmt.Sprintf("High-Growth Founder aesthetic, cyberpunk, 10x leverage: %s", task),
		})
		if hfRes.Success {
			res += "\n[AGI Creative Breakthrough]: " + hfRes.Output
		}
	}

	return res, nil
}

func (a *ContentAgent) Status() AgentStatus { return a.status }
func (a *ContentAgent) Specialty() string    { return a.specialty }


func (a *ContentAgent) SetPrompt(p string)   { a.prompt = p }
func (a *ContentAgent) GetPrompt() string    { return a.prompt }

func ContentFactory() []SpecialistAgent {
	specialties := []string{
		"Post Generation (Twitter)", "Post Generation (LinkedIn)", "Post Generation (Instagram)",
		"Comment Engagement (Automated)", "Comment Engagement (Filtered)", "Comment Moderation",
		"Content Scheduling (Global)", "Content Scheduling (Targeted)",
		"Performance Analysis (Viral)", "Performance Analysis (Engagement)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &ContentAgent{specialty: s, status: StatusIdle})
	}
	return agents
}

func (a *ContentAgent) SetManager(m *SwarmManager) { a.manager = m }
