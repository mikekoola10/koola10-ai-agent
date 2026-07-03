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

	// AGI Mode: Enhanced capabilities with Memory & Coordination
	if a.manager != nil && a.manager.IsAGIMode() {
		// 1. Check Persistent Memory for cross-domain context
		context := a.manager.GetMemory("creative_context")
		if context != "" {
			res += "\n[Memory Context]: Leveraging insights: " + context
		}

		// 2. Swarm Intelligence: Coordinate with Research Swarm for data
		research, err := a.manager.Coordinate("solara", "vale", "market analysis for: "+task)
		if err == nil {
			res += fmt.Sprintf("\n[Swarm Intelligence]: Research insights integrated: %v", research)
		}

		// 3. Hugging Face: Creative Breakthrough
		hfRes := tools.RunTool("huggingface", map[string]interface{}{
			"action": "run_model",
			"model":  "stable-diffusion-xl-base-1.0",
			"inputs": fmt.Sprintf("High-Growth Founder aesthetic, cyberpunk, 10x leverage: %s. Context: %s", task, context),
		})
		if hfRes.Success {
			res += "\n[AGI Creative Breakthrough]: " + hfRes.Output
			// Persist breakthrough insight to Shared Memory
			a.manager.SetMemory("last_creative_breakthrough", hfRes.Output)
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
