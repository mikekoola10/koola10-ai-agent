package agents

import (
	"fmt"
	"koola10/mirror"
	"os"
	"path/filepath"
)

type LeadGenAgent struct {
	specialty string
	status    AgentStatus
	mirror    *mirror.Mirror
}

func (a *LeadGenAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusIdle }()

	if a.mirror != nil {
		ctx := a.mirror.GetContext("leadgen")
		_ = ctx.Tone
	}

	// Simulate daily CSV production
	dir := "/data/leads/"
	os.MkdirAll(dir, 0755)
	filename := fmt.Sprintf("%s_leads.csv", a.specialty)
	content := "name,company,email,status\nJohn Doe,Acme Corp,john@acme.com,qualified"
	os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)

	a.status = StatusCompleted

	if a.mirror != nil {
		a.mirror.RecordOutcome("leadgen", map[string]interface{}{"task": task, "success": true})
	}

	return "Leads generated in " + filename, nil
}

func (a *LeadGenAgent) Status() AgentStatus { return a.status }
func (a *LeadGenAgent) Specialty() string    { return a.specialty }

func LeadGenFactory(m *mirror.Mirror) func() []SpecialistAgent {
	return func() []SpecialistAgent {
	specialties := []string{
		"LinkedIn Scraper (Tech)", "LinkedIn Scraper (Finance)", "LinkedIn Scraper (Healthcare)",
		"Crunchbase Enrichment", "Crunchbase Deep Dive",
		"Email Verification (Primary)", "Email Verification (Secondary)",
		"ICP Scoring", "Outreach Sequencing", "CRM Sync",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
		for _, s := range specialties {
			agents = append(agents, &LeadGenAgent{specialty: s, status: StatusIdle, mirror: m})
		}
		return agents
	}
}
