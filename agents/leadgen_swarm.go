package agents

import (
	"fmt"
	"os"
	"path/filepath"
)

type LeadGenAgent struct {
	specialty string
	status    AgentStatus
}

func (a *LeadGenAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking

	// Simulate daily CSV production
	dir := "/data/leads/"
	os.MkdirAll(dir, 0755)
	filename := fmt.Sprintf("%s_leads.csv", a.specialty)
	content := "name,company,email,status\nJohn Doe,Acme Corp,john@acme.com,qualified"
	os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)

	NotifyCelebration("new_lead", fmt.Sprintf("📬 New lead alert! Found some fresh leads via %s! The empire grows!", a.specialty))

	a.status = StatusCompleted
	return "Leads generated in " + filename, nil
}

func (a *LeadGenAgent) Status() AgentStatus { return a.status }
func (a *LeadGenAgent) Specialty() string    { return a.specialty }

func LeadGenFactory() []SpecialistAgent {
	specialties := []string{
		"LinkedIn Scraper (Tech)", "LinkedIn Scraper (Finance)", "LinkedIn Scraper (Healthcare)",
		"Crunchbase Enrichment", "Crunchbase Deep Dive",
		"Email Verification (Primary)", "Email Verification (Secondary)",
		"ICP Scoring", "Outreach Sequencing", "CRM Sync",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &LeadGenAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
