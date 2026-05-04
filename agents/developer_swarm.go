package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type DeveloperAgent struct {
	specialty string
	status    AgentStatus
	SendEmail func(to, subject, body string) error
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

		if nsTask.ReportTo != "" && a.SendEmail != nil {
			subject := fmt.Sprintf("Koola10 Night Shift Report - %s", time.Now().Format("2006-01-02"))
			body := fmt.Sprintf("Night Shift completed for repositories: %v\nTasks executed: %v\n\nPlease check the dashboard for the full collaboration summary.", nsTask.Repositories, nsTask.Tasks)
			a.SendEmail(nsTask.ReportTo, subject, body)
		}

		return map[string]interface{}{
			"status": "success",
			"message": fmt.Sprintf("Autonomous developer (%s) processed %d tasks across %d repositories", a.specialty, len(nsTask.Tasks), len(nsTask.Repositories)),
			"vertical": "night-shift",
			"report_sent_to": nsTask.ReportTo,
		}, nil
	}

	// Fallback for simple string tasks
	log.Printf("[DeveloperAgent] Processing simple task: %s", task)
	return fmt.Sprintf("Completed %s task: %s", a.specialty, task), nil
}

func (a *DeveloperAgent) Status() AgentStatus { return a.status }
func (a *DeveloperAgent) Specialty() string    { return a.specialty }

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
