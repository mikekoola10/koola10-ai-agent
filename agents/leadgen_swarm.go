package agents

import (
	"encoding/json"
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

	if task == "generate_nonprofit_prospects" {
		dir := "./data/leads/"
		os.MkdirAll(dir, 0755)
		filename := "prospects_nonprofit.json"

		type Prospect struct {
			Name    string  `json:"name"`
			Mission string  `json:"mission"`
			Budget  float64 `json:"budget"`
			Email   string  `json:"email"`
		}

		prospects := make([]Prospect, 50)
		for i := 0; i < 50; i++ {
			budget := 500000.0 + float64(i)*10000.0
			prospects[i] = Prospect{
				Name:    fmt.Sprintf("Nonprofit Organization %d", i+1),
				Mission: fmt.Sprintf("Mission to improve sector %d through innovation and community support.", i%10),
				Budget:  budget,
				Email:   fmt.Sprintf("contact@nonprofit%d.org", i+1),
			}
		}

		data, _ := json.MarshalIndent(prospects, "", "  ")
		os.WriteFile(filepath.Join(dir, filename), data, 0644)

		a.status = StatusCompleted
		return fmt.Sprintf("Generated 50 nonprofit prospects in %s", filename), nil
	}

	// Simulate daily CSV production
	dir := "./data/leads/"
	os.MkdirAll(dir, 0755)
	filename := fmt.Sprintf("%s_leads.csv", a.specialty)
	content := "name,company,email,status\nJohn Doe,Acme Corp,john@acme.com,qualified"
	os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)

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
