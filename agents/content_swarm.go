package agents

import (
	"fmt"
	"koola10/tools"
	"strings"
)

type ContentAgent struct {
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *ContentAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// High-leverage video generation via Hyperframes
	if strings.Contains(strings.ToLower(task), "video") || strings.Contains(strings.ToLower(task), "film") {
		skill := "general-video"
		if strings.Contains(strings.ToLower(task), "pr") {
			skill = "pr-to-video"
		} else if strings.Contains(strings.ToLower(task), "explainer") {
			skill = "faceless-explainer"
		} else if strings.Contains(strings.ToLower(task), "motion") {
			skill = "motion-graphics"
		}

		res := tools.RunTool("hyperframes", map[string]interface{}{
			"action":  "render",
			"project": "auto-content-" + a.specialty,
			"skill":   skill,
		})
		if res.Success {
			return res.Data, nil
		}
		// Log error but continue with fallback for simulation robustness
		fmt.Printf("Hyperframes failed: %s\n", res.Error)
	}

	// AI Image/Motion workflows via RunComfy
	if strings.Contains(strings.ToLower(task), "ai") || strings.Contains(strings.ToLower(task), "motion") {
		action := "generate"
		if strings.Contains(strings.ToLower(task), "outpaint") {
			action = "outpainting"
		} else if strings.Contains(strings.ToLower(task), "inpaint") {
			action = "inpainting"
		}

		res := tools.RunTool("runcomfy", map[string]interface{}{
			"action":   action,
			"workflow": task,
		})
		if res.Success {
			return res.Data, nil
		}
	}

	// Default simulated execution
	res := "Content Result (" + a.specialty + "): " + a.prompt + " | Task: " + task
	return res, nil
}

func (a *ContentAgent) Status() AgentStatus { return a.status }
func (a *ContentAgent) Specialty() string    { return a.specialty }
func (a *ContentAgent) SetPrompt(p string)   { a.prompt = p }
func (a *ContentAgent) GetPrompt() string    { return a.prompt }

func ContentFactory() []SpecialistAgent {
	specialties := []string{
		"Post Generation (Twitter)", "Post Generation (LinkedIn)", "Post Generation (Instagram)",
		"Hyperframes Video Architect", "RunComfy Motion Expert", "AI Avatar Producer",
		"Content Scheduling (Global)", "Content Scheduling (Targeted)",
		"Performance Analysis (Viral)", "Performance Analysis (Engagement)",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &ContentAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
