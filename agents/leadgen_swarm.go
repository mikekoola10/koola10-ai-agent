package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LeadGenAgent struct {
	ID        int
	Role      string
	Leads     int
	Revenue   float64
	Manager   *SwarmManager
}

func (a *LeadGenAgent) Specialty() string { return a.Role }
func (a *LeadGenAgent) Status() string    { return "active" }
func (a *LeadGenAgent) GetRevenue() float64 { return a.Revenue }

func (a *LeadGenAgent) Run(task string) string {
	if a.ID == 6 { // ICP Scoring Agent
		score, tokens := CallDeepSeek(task, "You are an ICP scoring agent. Rate the following lead potential from 1-100 and justify.")
		a.Manager.LedgerLogger("ai_inference", float64(tokens)*0.000002, "ICP scoring")
		return fmt.Sprintf("[ICP Agent] Scoring result: %s", score)
	}

	a.Leads += 5
	a.Revenue += 10.0 // $2 per lead

	msg := fmt.Sprintf("[%s] Generated 5 leads for task: %s", a.Role, task)

	// Write CSV export
	os.MkdirAll("/data/leads", 0755)
	timestamp := time.Now().Format("20060102_150405")
	csvPath := filepath.Join("/data/leads", fmt.Sprintf("leads_%d_%s.csv", a.ID, timestamp))
	data := fmt.Sprintf("name,email,industry\nLead_%d,lead%d@example.com,%s", a.ID, time.Now().Unix(), a.Role)
	os.WriteFile(csvPath, []byte(data), 0644)

	a.Manager.LedgerLogger("leadgen", 0.05, msg)
	a.Manager.AuditLogger("leads_generated", map[string]interface{}{
		"agent_id": a.ID,
		"role":     a.Role,
		"count":    5,
	})

	return msg
}

func GetLeadGenFactory(sm *SwarmManager) func(id int) SpecialistAgent {
	roles := []string{
		"LinkedIn Scraper (Tech)",
		"LinkedIn Scraper (Healthcare)",
		"LinkedIn Scraper (Finance)",
		"Crunchbase Enrichment",
		"Bulk Email Verification",
		"Individual Email Verification",
		"ICP Scoring (DeepSeek)",
		"Outreach Sequencing",
		"CRM Sync (HubSpot)",
		"Lead Analytics",
	}

	return func(id int) SpecialistAgent {
		return &LeadGenAgent{
			ID:      id,
			Role:    roles[id%len(roles)],
			Manager: sm,
		}
	}
}
