package agents

import (
	"fmt"
	"koola10/tools"
	"strings"
)

type GrantSwarmAgent struct {
	specialty string
	status    AgentStatus
	prompt    string
}

func (a *GrantSwarmAgent) Run(task string) (interface{}, error) {
	a.status = StatusWorking
	defer func() { a.status = StatusCompleted }()

	// Use Granted MCP tools for relevant tasks
	if strings.Contains(strings.ToLower(task), "search") || strings.Contains(strings.ToLower(task), "find") {
		res := tools.RunTool("search_grants", map[string]interface{}{
			"query": task,
		})
		if res.Success {
			return res.Data, nil
		}
		return nil, fmt.Errorf("search_grants tool failed: %s", res.Error)
	}

	if strings.Contains(strings.ToLower(task), "details") || strings.Contains(strings.ToLower(task), "get") {
		// Try to extract slug or id from task
		res := tools.RunTool("get_grant", map[string]interface{}{
			"slug": task,
		})
		if res.Success {
			return res.Data, nil
		}
		// Fallback to searching funders if it looks like funder research
		if strings.Contains(strings.ToLower(task), "funder") || strings.Contains(strings.ToLower(task), "foundation") {
			res = tools.RunTool("search_funders", map[string]interface{}{
				"query": task,
			})
			if res.Success {
				return res.Data, nil
			}
		}
	}

	// Default simulated execution if no tool matches or fails
	res := "Grant Proposal (" + a.specialty + "): " + a.prompt + " | Task: " + task
	return res, nil
}

func (a *GrantSwarmAgent) Status() AgentStatus { return a.status }
func (a *GrantSwarmAgent) Specialty() string    { return a.specialty }
func (a *GrantSwarmAgent) SetPrompt(p string)   { a.prompt = p }
func (a *GrantSwarmAgent) GetPrompt() string    { return a.prompt }

func GrantSwarmFactory() []SpecialistAgent {
	specialties := []string{
		"Federal Database Monitor", "Federal Proposal Draft", "Federal Compliance",
		"State Grant Search", "State Proposal Draft", "State Budget Plan",
		"Foundation Outreach", "Foundation Proposal", "Private Grant Search", "Impact Report Gen",
	}
	agents := make([]SpecialistAgent, 0, len(specialties))
	for _, s := range specialties {
		agents = append(agents, &GrantSwarmAgent{specialty: s, status: StatusIdle})
	}
	return agents
}
