package agents

import (
	"fmt"
	"os"
	"path/filepath"
)

type LeadGenAgent struct {
	manager *SwarmManager
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *LeadGenAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking

	// Simulate daily CSV production
	dir := "/data/leads/"
	os.MkdirAll(dir, 0755)
	filename := fmt.Sprintf("%s_leads.csv", a.specialty)
	content := fmt.Sprintf("name,company,email,status,prompt\nJohn Doe,Acme Corp,john@acme.com,qualified,%s", a.prompt)
	os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)

	a.status = StatusCompleted
	return "Leads generated in " + filename + " with prompt context", nil
}

func (a *LeadGenAgent) Status() AgentStatus { return a.status }
func (a *LeadGenAgent) Specialty() string    { return a.specialty }
func (a *LeadGenAgent) SetPrompt(p string)   { a.prompt = p }
func (a *LeadGenAgent) GetPrompt() string    { return a.prompt }

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

func (a *LeadGenAgent) SetManager(m *SwarmManager) { a.manager = m }

func (a *LeadGenAgent) ConfidenceLevel() float64 { return 0.95 }
func (a *LeadGenAgent) RequestClarification(ctx string) string { return "" }
