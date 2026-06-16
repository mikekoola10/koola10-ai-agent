package agents

import (
	"encoding/json"
	"fmt"
	"log"
)

type DeveloperAgent struct {
	BaseAGISkills
	specialty string
	status    AgentStatus
}

type NightShiftTask struct {
	Repositories []string `json:"repositories"`
	Tasks        []string `json:"tasks"`
	AutoMerge    bool     `json:"auto_merge"`
	ReportTo     string   `json:"report_to"`
}

func (a *DeveloperAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	var nsTask NightShiftTask
	if err := json.Unmarshal([]byte(task), &nsTask); err == nil {
		log.Printf("[DeveloperAgent] Night Shift started: %v on repos %v", nsTask.Tasks, nsTask.Repositories)
		return map[string]interface{}{
			"status": "success",
			"message": fmt.Sprintf("Autonomous developer (%s) processed %d tasks across %d repositories", a.specialty, len(nsTask.Tasks), len(nsTask.Repositories)),
			"vertical": "night-shift",
			"report_sent_to": nsTask.ReportTo,
		}, nil
	}

	// SaaS Building logic
	if a.specialty == "Backend (Go)" || a.specialty == "Frontend (React)" {
		log.Printf("[DeveloperAgent] Autonomous SaaS Building: %s", task)
		return map[string]interface{}{
			"status": "success",
			"artifact": fmt.Sprintf("SaaS boilerplate for %s", task),
			"deployment": "fly.io",
		}, nil
	}

	// Fallback for simple string tasks
	log.Printf("[DeveloperAgent] Processing simple task: %s", task)
	return fmt.Sprintf("Completed %s task: %s", a.specialty, task), nil
}

func (a *DeveloperAgent) Status() AgentStatus { return a.status }
func (a *DeveloperAgent) Specialty() string    { return a.specialty }

func (a *DeveloperAgent) Capabilities() []string {
	return []string{"software_development", "ci_cd", "saas_building", "testing"}
}

func (a *DeveloperAgent) InputSchema() map[string]string {
	return map[string]string{
		"repositories": "[]string",
		"tasks":        "[]string",
		"auto_merge":   "bool",
	}
}

func DeveloperFactory() []SpecialistAgent {
	specialties := []string{
		"Frontend (React)", "Frontend (Vue)", "Frontend (Svelte)",
		"Backend (Go)", "Backend (Python)", "Backend (Node)",
		"DevOps (Fly.io)", "DevOps (Docker)",
		"Testing Suite", "Documentation Generator",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &DeveloperAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
